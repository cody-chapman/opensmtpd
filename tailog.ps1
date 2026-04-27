#!/usr/bin/pwsh

# --- Configuration ---
$Port = 8080
$Hostname = hostname
$Logs = @{
    "SMTP"   = "/var/log/opensmtpd/opensmtpd.log"
    "PWSH"   = "/var/log/opensmtpd/pwsh.log"
    "SUPERVISOR"    = "/var/log/supervisor/supervisord.log"
}

# --- 2. START SERVER ---
$Listener = New-Object System.Net.HttpListener
$Listener.Prefixes.Add("http://*:$Port/")
try { 
    $Listener.Start() 
    Write-Host "===============================================" -ForegroundColor Cyan
    Write-Host " DASHBOARD: http://localhost:$Port" -ForegroundColor Green
    Write-Host "===============================================" -ForegroundColor Cyan
} catch { 
    Write-Error "Start failed. Ensure you are running as sudo."; exit 
}

while ($Listener.IsListening) {
    $Context = $Listener.GetContext()
    $Request = $Context.Request
    $Response = $Context.Response

    # --- ROUTE: UI ---
    if ($Request.Url.LocalPath -eq "/") {
        $Response.ContentType = "text/html"
        $Buttons = ""; foreach($k in ($Logs.Keys | Sort-Object)) { 
            $Buttons += "<button class='filter-btn active' data-source='$k'>$k</button>" 
        }
        $Html = @"
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { background: #0d1117; color: #c9d1d9; font-family: 'Consolas', monospace; margin: 0; display: flex; flex-direction: column; height: 100vh; }
        header { background: #161b22; padding: 15px; border-bottom: 1px solid #30363d; display: flex; align-items: center; gap: 15px; }
        .brand { color: #3fb950; font-weight: bold; font-size: 1.2em; margin-right: 20px; }
        .filter-btn { background: #21262d; border: 1px solid #30363d; color: #c9d1d9; padding: 6px 12px; border-radius: 6px; cursor: pointer; }
        .filter-btn.active { background: #238636; border-color: #3fb950; }
        #log-container { flex: 1; overflow-y: auto; padding: 20px; background: #010409; }
        .line { margin-bottom: 4px; font-size: 13px; display: flex; gap: 10px; border-bottom: 1px solid #161b22; padding-bottom: 2px; }
        .line.hidden { display: none; }
        .ts { color: #8b949e; min-width: 80px; }
        .badge { font-weight: bold; padding: 2px 8px; border-radius: 4px; font-size: 0.8em; min-width: 60px; text-align: center; }
        .tag-SMTP { background: #1f6feb; } .tag-AUTH { background: #da3633; } .tag-SYS { background: #8957e5; } .tag-INIT { background: #6e7681; }
    </style>
</head>
<body>
    <header>
        <div class="brand">SYSTEM LOGS</div>
        <div class="filters">$Buttons</div>
    </header>
    <div id="log-container"><div id="output"></div></div>
    <script>
        const output = document.getElementById('output');
        const container = document.getElementById('log-container');
        const activeFilters = new Set();
        
        document.querySelectorAll('.filter-btn').forEach(btn => {
            activeFilters.add(btn.dataset.source);
            btn.onclick = () => {
                const src = btn.dataset.source;
                if(activeFilters.has(src)) { activeFilters.delete(src); btn.classList.remove('active'); }
                else { activeFilters.add(src); btn.classList.add('active'); }
                document.querySelectorAll('.line').forEach(l => l.classList.toggle('hidden', !activeFilters.has(l.dataset.source)));
            };
        });

        const es = new EventSource('/stream');
        es.onmessage = (e) => {
            const parts = e.data.split('|');
            if (parts.length < 3) return;
            const [src, ts, ...msg] = parts;
            const div = document.createElement('div');
            div.className = 'line';
            div.dataset.source = src;
            div.innerHTML = '<span class="ts">'+ts+'</span><span class="badge tag-'+src+'">'+src+'</span><span>'+msg.join('|')+'</span>';
            output.appendChild(div);
            container.scrollTop = container.scrollHeight;
        };
        es.onerror = () => { console.log("Reconnecting..."); };
    </script>
</body></html>
"@
        $Buffer = [System.Text.Encoding]::UTF8.GetBytes($Html)
        $Response.OutputStream.Write($Buffer, 0, $Buffer.Length)
        $Response.Close()
    } 

    # --- ROUTE: STREAM ---
    elseif ($Request.Url.LocalPath -eq "/stream") {
        $Response.ContentType = "text/event-stream"
        $Response.Headers.Add("Cache-Control", "no-cache")
        $Response.Headers.Add("Connection", "keep-alive")
        $Writer = New-Object System.IO.StreamWriter($Response.OutputStream)

        Write-Host "[CONN] Client joined the stream." -ForegroundColor Cyan

        # Send an immediate "System Ready" line so the user knows it's working
        $now = Get-Date -Format "HH:mm:ss"
        $Writer.WriteLine("data: INIT|$now|Stream connection established. Waiting for logs...")
        $Writer.WriteLine(); $Writer.Flush()

        # We use a single 'tail' command to watch all files at once
        $ValidFiles = $Logs.Values | Where-Object { Test-Path $_ }
        
        if ($ValidFiles.Count -eq 0) {
            $Writer.WriteLine("data: SYS|$now|ERROR: No log files found!")
            $Writer.Flush(); $Response.Close(); continue
        }

        # Start 'tail' as a background process
        $ProcInfo = New-Object System.Diagnostics.ProcessStartInfo
        $ProcInfo.FileName = "tail"
        # -f = follow, -n 5 = show last 5 lines, -v = always show headers (used to identify source)
        $ProcInfo.Arguments = "-f -n 5 -v " + ($ValidFiles -join " ")
        $ProcInfo.RedirectStandardOutput = $true
        $ProcInfo.UseShellExecute = $false
        $Process = [System.Diagnostics.Process]::Start($ProcInfo)

        try {
            $CurrentSource = "LOG"
            while (!$Process.StandardOutput.EndOfStream) {
                $rawLine = $Process.StandardOutput.ReadLine()
                
                # 'tail -v' outputs "==> filename <==" when switching files
                if ($rawLine -match "==> (.*) <==") {
                    $fileName = $Matches[1]
                    # Map the filename back to our Key (SMTP, AUTH, etc.)
                    foreach($pair in $Logs.GetEnumerator()) {
                        if ($fileName -like "*$($pair.Value)*") { $CurrentSource = $pair.Key; break }
                    }
                    continue # Don't print the header line itself
                }

                if ($rawLine.Trim()) {
                    $ts = Get-Date -Format "HH:mm:ss"
                    $Writer.WriteLine("data: $CurrentSource|$ts|$rawLine")
                    $Writer.WriteLine()
                    $Writer.Flush()
                }
            }
        } catch {
            Write-Host "[DISC] Client disconnected." -ForegroundColor Gray
        } finally {
            if (!$Process.HasExited) { $Process.Kill() }
            $Response.Close()
        }
    }
}
