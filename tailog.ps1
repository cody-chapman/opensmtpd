#!/usr/bin/pwsh

# --- Configuration ---
$Port = 8080
$Hostname = hostname
$Logs = @{
    "SMTP"   = "/var/log/opensmptd/opensmtpd.log"
    "PWSH"   = "/var/log/opensmtpd/pwsh.log"
    "SUPERVISOR"    = "/var/log/supervisor/supervisord.log"
}

# --- Initialization ---
$Listener = New-Object System.Net.HttpListener
$Listener.Prefixes.Add("http://*:$Port/")
try { $Listener.Start() } catch { 
    Write-Error "Start failed. Run as sudo? Port $Port free?"; exit 
}

Write-Host "Dashboard active at http://localhost:$Port" -ForegroundColor Cyan

try {
    while ($Listener.IsListening) {
        $Context = $Listener.GetContext()
        $Request = $Context.Request
        $Response = $Context.Response

        if ($Request.Url.LocalPath -eq "/") {
            $Response.ContentType = "text/html"
            
            # Create Filter Buttons dynamically based on our Hashtable
            $Buttons = ""
            foreach($key in $Logs.Keys) {
                $Buttons += "<button class='filter-btn active' data-source='$key'>$key</button>"
            }

            $Html = @"
<!DOCTYPE html>
<html>
<head>
    <title>$Hostname Logs</title>
    <style>
        :root {
            --bg: #0d1117; --panel: #161b22; --border: #30363d;
            --text: #c9d1d9; --dim: #8b949e; --green: #238636;
        }
        body { background: var(--bg); color: var(--text); font-family: monospace; margin: 0; display: flex; flex-direction: column; height: 100vh; }
        
        header { 
            background: var(--panel); padding: 15px 20px; border-bottom: 1px solid var(--border);
            display: flex; align-items: center; gap: 20px;
        }

        /* Filter UI */
        .filters { display: flex; gap: 10px; }
        .filter-btn { 
            background: #21262d; border: 1px solid var(--border); color: var(--text);
            padding: 5px 15px; border-radius: 6px; cursor: pointer; transition: 0.2s;
        }
        .filter-btn.active { background: var(--green); border-color: #3fb950; }
        .filter-btn:not(.active) { opacity: 0.5; }

        #log-container { flex: 1; overflow-y: auto; padding: 15px; }
        .line { margin-bottom: 4px; display: block; border-left: 3px solid transparent; padding-left: 10px; }
        .line.hidden { display: none; }
        
        /* Source Badges */
        .badge { font-weight: bold; padding: 2px 6px; border-radius: 3px; font-size: 0.85em; margin-right: 10px; display: inline-block; width: 60px; text-align: center; }
        .tag-SMTP { background: #1f6feb; color: white; }
        .tag-AUTH { background: #da3633; color: white; }
        .tag-SYS { background: #8957e5; color: white; }
        .tag-KERNEL { background: #d29922; color: white; }

        .ts { color: var(--dim); margin-right: 10px; }
    </style>
</head>
<body>
    <header>
        <div style="font-weight:bold; color:var(--green)">$Hostname :: LOGS</div>
        <div class="filters">
            $Buttons
            <button onclick="clearLogs()" style="margin-left:20px" class="filter-btn">Clear</button>
        </div>
    </header>
    <div id="log-container"><div id="output"></div></div>

    <script>
        const output = document.getElementById('output');
        const container = document.getElementById('log-container');
        const activeFilters = new Set(JSON.parse('$($Logs.Keys | ConvertTo-Json -Compress)'));

        // Toggle Filters
        document.querySelectorAll('.filter-btn[data-source]').forEach(btn => {
            btn.onclick = () => {
                const src = btn.dataset.source;
                if (activeFilters.has(src)) {
                    activeFilters.delete(src);
                    btn.classList.remove('active');
                } else {
                    activeFilters.add(src);
                    btn.classList.add('active');
                }
                updateVisibility();
            };
        });

        function updateVisibility() {
            document.querySelectorAll('.line').forEach(line => {
                if (activeFilters.has(line.dataset.source)) {
                    line.classList.remove('hidden');
                } else {
                    line.classList.add('hidden');
                }
            });
        }

        function clearLogs() { output.innerHTML = ''; }

        const es = new EventSource('/stream');
        es.onmessage = (e) => {
            // Data format: SOURCE|TIMESTAMP|MESSAGE
            const parts = e.data.split('|');
            if (parts.length < 3) return;

            const [src, ts, msg] = parts;
            const div = document.createElement('div');
            div.className = 'line';
            div.dataset.source = src;
            if (!activeFilters.has(src)) div.classList.add('hidden');

            div.innerHTML = `<span class="ts">` + ts + `</span>` +
                           `<span class="badge tag-` + src + `">` + src + `</span>` +
                           `<span>` + msg + `</span>`;
            
            output.appendChild(div);
            container.scrollTop = container.scrollHeight;
        };
    </script>
</body>
</html>
"@
            $Buffer = [System.Text.Encoding]::UTF8.GetBytes($Html)
            $Response.OutputStream.Write($Buffer, 0, $Buffer.Length)
            $Response.Close()
        }

elseif ($Request.Url.LocalPath -eq "/stream") {
    $Response.ContentType = "text/event-stream"
    $Response.Headers.Add("Cache-Control", "no-cache")
    $Response.Headers.Add("Connection", "keep-alive")
    
    $Writer = New-Object System.IO.StreamWriter($Response.OutputStream)

    # 1. Verify files exist before trying to tail them
    $ExistingLogs = @{}
    foreach($key in $Logs.Keys) {
        if (Test-Path $Logs[$key]) {
            $ExistingLogs[$key] = $Logs[$key]
            Write-Host "[DEBUG] Watching $key -> $($Logs[$key])" -ForegroundColor Green
        } else {
            Write-Host "[DEBUG] Skipping $key: File not found at $($Logs[$key])" -ForegroundColor Yellow
        }
    }

    if ($ExistingLogs.Count -eq 0) {
        $Writer.WriteLine("data: SYS|$(Get-Date)|ERROR: No log files found or accessible.")
        $Writer.Flush(); $Response.Close(); continue
    }

    # 2. Start the stream
    try {
        # Using -Tail 1 ensures you see SOMETHING immediately on connect
        Get-Content -Path $ExistingLogs.Values -Tail 1 -Wait -ErrorAction SilentlyContinue | ForEach-Object {
            $rawLine = $_
            
            # Find which log this line belongs to
            # We use the filename in the path as the primary check
            $foundSrc = "LOG"
            foreach($key in $ExistingLogs.Keys) {
                $fileName = Split-Path $ExistingLogs[$key] -Leaf
                if ($_.PSPath -like "*$fileName*") { $foundSrc = $key; break }
            }

            $ts = [DateTime]::Now.ToString("HH:mm:ss")
            $dataString = "$foundSrc|$ts|$rawLine"

            try {
                $Writer.WriteLine("data: $dataString")
                $Writer.WriteLine()
                $Writer.Flush()
                # Debug output to your terminal so you can see it's working
                Write-Host "[SEND] $dataString" -ForegroundColor Gray
            } catch { 
                Write-Host "[DEBUG] Client disconnected." -ForegroundColor Red
                break 
            }
        }
    } finally {
        $Response.Close()
    }
}
