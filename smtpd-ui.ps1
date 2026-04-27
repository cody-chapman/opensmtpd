Import-Module Pode

Start-PodeServer {
    # 1. Setup Endpoint
    Add-PodeEndpoint -Address * -Port 8080 -Protocol Http

    # 2. Status Helper
    function Get-SmtpStatus {
        try {
            $stats = & smtpctl show stats 2>$null
            if ($lastExitCode -eq 0) { 
                return @{ State = "RUNNING"; Color = "#00ff88"; Stats = ($stats -join "`n") } 
            }
        } catch {}
        return @{ State = "STOPPED"; Color = "#ff4444"; Stats = "Daemon is not reachable." }
    }

    # 3. Main Dashboard Route
    Add-PodeRoute -Method Get -Path "/" -ScriptBlock {
        $status = Get-SmtpStatus
        $queryOut = $WebEvent.Query['out'] 
        $displayOut = if ($queryOut) { [uri]::UnescapeDataString($queryOut) } else { "Waiting for command..." }

        # Prepare Data for the template
        $viewData = @{
            Host   = $env:COMPUTERNAME
            Time   = (Get-Date -Format "HH:mm:ss")
            State  = $status.State
            Color  = $status.Color
            Stats  = $status.Stats
            Output = $displayOut
        }

        # Build HTML
        $html = @"
<!DOCTYPE html>
<html>
<head>
    <title>OpenSMTPD Manager</title>
    <style>
        body { background: #0d1117; color: #c9d1d9; font-family: sans-serif; padding: 40px; margin: 0; }
        .container { max-width: 800px; margin: auto; }
        .header { display: flex; justify-content: space-between; border-bottom: 1px solid #30363d; padding-bottom: 10px; }
        .card { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 20px; margin-top: 20px; }
        .btn { background: #21262d; color: #58a6ff; border: 1px solid #30363d; padding: 8px 15px; border-radius: 6px; cursor: pointer; font-weight: bold; }
        .btn:hover { background: #30363d; }
        .btn-run { background: #238636; color: white; border: none; }
        .btn-stop { background: #f85149; color: white; border: none; }
        pre { background: #000; padding: 15px; border-radius: 6px; border: 1px solid #30363d; color: #8b949e; overflow-x: auto; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div>
                <h1 style="margin:0;">OpenSMTPD Manager</h1>
                <small>$($viewData.Host) | $($viewData.Time)</small>
            </div>
            <div style="color: $($viewData.Color); font-weight: bold;">$($viewData.State)</div>
        </div>

        <div class="card">
            <h3>Actions</h3>
            <form action="/action/start" method="post" style="display:inline;"><button class="btn btn-run">Start</button></form>
            <form action="/action/stop" method="post" style="display:inline;"><button class="btn btn-stop">Stop</button></form>
            <form action="/action/reload" method="post" style="display:inline;"><button class="btn">Reload</button></form>
            <form action="/action/flush" method="post" style="display:inline;"><button class="btn">Flush</button></form>
        </div>

        <div class="card">
            <h3>Stats</h3>
            <pre>$($viewData.Stats)</pre>
        </div>

        <div class="card">
            <h3>Console Output</h3>
            <pre style="color: #58a6ff;">$($viewData.Output)</pre>
        </div>
    </div>
</body>
</html>
"@
        Write-PodeWebResponse -Value $html -ContentType 'text/html'
    }

    # 4. Action Handlers
    Add-PodeRoute -Method Post -Path "/action/:act" -ScriptBlock {
        $action = $WebEvent.Parameters['act']
        $cmd = switch ($action) {
            "start"  { "start" }
            "stop"   { "stop" }
            "reload" { "reload" }
            "flush"  { "schedule all" }
        }

        $result = & smtpctl $cmd 2>&1 | Out-String
        Move-PodeResponse -Url "/?out=$([uri]::EscapeDataString($result))"
    }
}
