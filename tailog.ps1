#!/usr/bin/pwsh

# --- Configuration ---
$Port = 8080
$Hostname = hostname
$Logs = @{
    "SMTP"   = "/var/log/opensmptd/opensmtpd.log"
    "PWSH"   = "/var/log/opensmtpd/pwsh.log"
    "SUPERVISOR"    = "/var/log/supervisor/supervisord.log"
}

# --- 2. INITIALIZATION ---
$Listener = New-Object System.Net.HttpListener
$Listener.Prefixes.Add("http://*:$Port/")

try {
    $Listener.Start()
    Write-Host "------------------------------------------------" -ForegroundColor Cyan
    Write-Host " DASHBOARD ACTIVE: http://localhost:$Port" -ForegroundColor Green
    Write-Host " RUNNING AS: $(whoami)" -ForegroundColor Yellow
    Write-Host "------------------------------------------------" -ForegroundColor Cyan
} catch {
    Write-Error "Could not start server. Try: sudo pwsh logview.ps1"
    exit
}

# --- 3. MAIN SERVER LOOP ---
try {
    while ($Listener.IsListening) {
        $Context = $Listener.GetContext()
        $Request = $Context.Request
        $Response = $Context.Response

        # ROUTE: MAIN UI
        if ($Request.Url.LocalPath -eq "/") {
            $Response.ContentType = "text/html"
            
            # Dynamically build filter buttons
            $Buttons = ""
            foreach($key in ($Logs.Keys | Sort-Object)) {
                $Buttons += "<button class='filter-btn active' data-source='$key'>$key</button>"
            }

            $Html = @"
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>$Hostname Logs</title>
    <style>
        :root {
            --bg: #0d1117; --panel: #161b22; --border: #30363d;
            --text: #c9d1d9; --dim: #8b949e; --green: #3fb950; --red: #f85149;
        }
        body { background: var(--bg); color: var(--text); font-family: 'Cascadia Code', 'Courier New', monospace; margin: 0; display: flex; flex-direction: column; height: 100vh; overflow: hidden; }
        
        header { 
            background: var(--panel); padding: 12px 20px; border-bottom: 1px solid var(--border);
            display: flex; align-items: center; justify-content: space-between;
        }

        .brand { font-weight: bold; color: var(--green); display: flex; align-items: center; gap: 10px; }
        .pulse { width: 10px; height: 10px; background: var(--green); border-radius: 50%; animation: blink 1.5s infinite; }
        @keyframes blink { 0% { opacity: 1; } 50% { opacity: 0.3; } 100% { opacity: 1; } }

        .filters { display: flex; gap: 8px; align-items: center; }
        .filter-btn { 
            background: #21262d; border: 1px solid var(--border); color: var(--text);
            padding: 4px 12px; border-radius: 6px; cursor: pointer; font-size: 12px;
        }
        .filter-btn.active { background: #238636; border-color: var(--green); }
        .filter-btn:not(.active) { opacity: 0.4; border-style: dashed; }
        .clear-btn { background: #da3633; border: none; color: white; margin-left: 15px; }

        #log-container { flex: 1; overflow-y: auto; padding: 15px; background: #010409; }
        .line { font-size: 13px; line-height: 1.5; margin-bottom: 2px; white-space: pre-wrap; display: flex; align-items: flex-start; }
        .line.hidden { display: none; }
        
        .ts { color: var(--dim); min-width: 80px; }
        .badge { font-size: 10px; font-weight: bold; padding: 1px 6px; border-radius: 3px; margin-right: 10px; width: 50px; text-align: center; text-transform: uppercase; }
        
        /* Dynamic Source Colors */
        .tag-SMTP { background: #1f6feb; color: white; }
        .tag-AUTH { background: #da3633; color: white; }
        .tag-SYS { background: #8957e5; color: white; }
        .tag-KERNEL { background: #d29922; color: white; }
        .tag-LOG { background: #6e7681; color: white; }

        .msg { word-break: break-all; }
    </style>
</head>
<body>
    <header>
        <div class="brand"><div class="pulse"></div> $Hostname LOG STREAM</div>
        <div class="filters">
            $Buttons
            <button onclick="clearLogs()" class="filter-btn clear-btn">Clear View</button>
        </div>
    </header>
    <div id="log-container"><div id="output"></div></div>

    <script>
        const output = document.getElementById('output');
        const container = document.getElementById('log-container');
        const activeFilters = new Set();
        
        // Initialize filters as active
        document.querySelectorAll('.filter-btn[data-source]').forEach(btn => {
            activeFilters.add(btn.dataset.source);
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
                line.classList.toggle('hidden', !activeFilters.has(line.dataset.source));
            });
        }

        function clearLogs() { output.innerHTML = ''; }

        function connect() {
            const es = new EventSource('/stream');
            
            es.onmessage = (e) => {
                const parts = e.data.split('|');
                if (parts.length < 3) return;

                const [src, ts, ...msgParts] = parts;
                const msg = msgParts.join('|'); // Rejoin in case message had pipes

                const div = document.createElement('div');
                div.className = 'line';
                div.dataset.source = src;
                if (!activeFilters.has(src)) div.classList.add('hidden');

                div.innerHTML = `<span class="ts">` + ts + `</span>` +
                               `<span class="badge tag-` + src + `">` + src + `</span>` +
                               `<span class="msg">` + msg + `</span>`;
                
                output.appendChild(div);
                
                // Only scroll if we are near the bottom
                const isScrolledToBottom = container.scrollHeight - container.clientHeight <= container.scrollTop + 50;
                if (isScrolledToBottom) {
                    container.scrollTop = container.scrollHeight;
                }
            };

            es.onerror = () => {
                es.close();
                setTimeout(connect, 2000); // Reconnect loop
            };
        }

        connect();
    </script>
</body>
</html>
"@
            $Buffer = [System.Text.Encoding]::UTF8.GetBytes($Html)
            $Response.OutputStream.Write($Buffer, 0, $Buffer.Length)
            $Response.Close()
        }

        # ROUTE: THE STREAM (SSE)
        elseif ($Request.Url.LocalPath -eq "/stream") {
            $Response.ContentType = "text/event-stream"
            $Response.Headers.Add("Cache-Control", "no-cache")
            $Response.Headers.Add("Connection", "keep-alive")
            $Writer = New-Object System.IO.StreamWriter($Response.OutputStream)

            # Filter out non-existent files to prevent pipeline errors
            $ExistingPaths = @()
            $PathToKey = @{}
            foreach($key in $Logs.Keys) {
                if (Test-Path $Logs[$key]) {
                    $ExistingPaths += $Logs[$key]
                    # Map the actual leaf name to our Key
                    $leaf = Split-Path $Logs[$key] -Leaf
                    $PathToKey[$leaf] = $key
                }
            }

            if ($ExistingPaths.Count -eq 0) {
                $Writer.WriteLine("data: LOG|ERROR|No log files found at specified paths.")
                $Writer.Flush(); $Response.Close(); continue
            }

            Write-Host "[INFO] Client connected. Streaming $($ExistingPaths.Count) files." -ForegroundColor Cyan

            try {
                # -Tail 1 ensures the connection stays open even if logs are quiet
                Get-Content -Path $ExistingPaths -Tail 1 -Wait -ErrorAction SilentlyContinue | ForEach-Object {
                    $line = $_
                    $ts = [DateTime]::Now.ToString("HH:mm:ss")
                    
                    # Identify the source by checking the PSPath property
                    $foundKey = "LOG"
                    foreach($leaf in $PathToKey.Keys) {
                        if ($line.PSPath -like "*$leaf*") {
                            $foundKey = $PathToKey[$leaf]
                            break
                        }
                    }

                    try {
                        $Writer.WriteLine("data: $foundKey|$ts|$line")
                        $Writer.WriteLine()
                        $Writer.Flush()
                    } catch {
                        throw "Client Disconnected"
                    }
                }
            } catch {
                Write-Host "[INFO] Client disconnected." -ForegroundColor Gray
            } finally {
                $Response.Close()
            }
        }
        else {
            $Response.StatusCode = 404
            $Response.Close()
        }
    }
} finally {
    $Listener.Stop()
}
