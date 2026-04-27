Import-Module Pode

Start-PodeServer {
    # Bind to port 8080 (Change to 8888 if 8080 is busy)
    Add-PodeEndpoint -Address * -Port 8888 -Protocol Http

    # ── Dashboard Route ──────────────────────────────────────────────────────
    Add-PodeRoute -Method Get -Path "/" -ScriptBlock {
        try {
            # 1. Capture Output from Redirects
            $queryOut = $WebEvent.Query['out']
            $displayOut = if ($queryOut) { [uri]::UnescapeDataString($queryOut) } else { "System Ready. Waiting for input..." }

            # 2. Get Daemon Status Safely
            $state = "UNKNOWN"
            $color = "#8b949e"
            $stats = "No stats available."

            try {
                $rawStats = & /usr/sbin/smtpctl show stats 2>$null
                if ($lastExitCode -eq 0) {
                    $state = "RUNNING"
                    $color = "#00ff88"
                    $stats = $rawStats -join "`n"
                } else {
                    $state = "STOPPED"
                    $color = "#f85149"
                }
            } catch {
                $state = "NOT FOUND"
                $stats = "Error: '/usr/sbin/smtpctl' command not found or permission denied."
            }

            # 3. Build the Dark Theme HTML
            $html = @"
<!DOCTYPE html>
<html>
<head>
    <title>OpenSMTPD Manager</title>
    <style>
        body { background: #0d1117; color: #c9d1d9; font-family: -apple-system, system-ui, sans-serif; padding: 40px; margin: 0; }
        .container { max-width: 900px; margin: auto; }
        .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #30363d; padding-bottom: 15px; }
        .card { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 20px; margin-top: 20px; }
        .status { color: $color; font-weight: bold; text-transform: uppercase; }
        .btn-group { display: flex; gap: 10px; margin-top: 10px; }
        .btn { background: #21262d; color: #58a6ff; border: 1px solid #30363d; padding: 10px 20px; border-radius: 6px; cursor: pointer; font-weight: bold; }
        .btn:hover { background: #30363d; }
        .btn-start { background: #238636; color: white; border: none; }
        .btn-stop { background: #f85149; color: white; border: none; }
        pre { background: #000; padding: 15px; border-radius: 6px; border: 1px solid #30363d; color: #8b949e; overflow-x: auto; white-space: pre-wrap; font-family: monospace; }
        .console { color: #58a6ff; border-color: #58a6ff33; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div>
                <h1 style="margin:0;">OpenSMTPD Manager</h1>
                <small style="color:#8b949e;">Server: $($env:COMPUTERNAME) | $(Get-Date -Format "HH:mm:ss")</small>
            </div>
            <div class="status">$state</div>
        </div>

        <div class="card">
            <h3 style="margin-top:0;">Service Controls</h3>
            <div class="btn-group">
                <form action="/action" method="post"><button class="btn btn-start" name="cmd" value="start">Start</button></form>
                <form action="/action" method="post"><button class="btn btn-stop" name="cmd" value="stop">Stop</button></form>
                <form action="/action" method="post"><button class="btn" name="cmd" value="reload">Reload</button></form>
                <form action="/action" method="post"><button class="btn" name="cmd" value="schedule all">Flush Queue</button></form>
            </div>
        </div>

        <div class="card">
            <h3 style="margin-top:0;">Runtime Statistics</h3>
            <pre>$stats</pre>
        </div>

        <div class="card">
            <h3 style="margin-top:0;">Console Output</h3>
            <pre class="console">$displayOut</pre>
        </div>
    </div>
</body>
</html>
"@
            Write-PodeWebResponse -Value $html -ContentType 'text/html'

        } catch {
            # If the Route itself fails, output the error as a string instead of a 500 page
            Write-PodeWebResponse -Value "Critical Route Error: $($_.Exception.Message)" -StatusCode 500
        }
    }

    # ── Action Handler ───────────────────────────────────────────────────────
    Add-PodeRoute -Method Post -Path "/action" -ScriptBlock {
        try {
            $cmd = $WebEvent.Data['cmd']
            if (-not $cmd) { throw "No command provided." }

            # Execute command and capture string output
            $result = & /usr/sbin/smtpctl $cmd 2>&1 | Out-String
            
            if ([string]::IsNullOrWhiteSpace($result)) {
                $result = "Success: Command '$cmd' executed (no output returned)."
            }

            $escaped = [uri]::EscapeDataString($result.Trim())
            Move-PodeResponse -Url "/?out=$escaped"

        } catch {
            $err = [uri]::EscapeDataString("Action Error: $($_.Exception.Message)")
            Move-PodeResponse -Url "/?out=$err"
        }
    }
}
