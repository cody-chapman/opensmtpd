Import-Module Pode

Start-PodeServer {
    # 1. Port Check: Ensure this port is definitely open
    Add-PodeEndpoint -Address * -Port 8080 -Protocol Http

    # 2. Define the path clearly at the top
    $SmtpPath = "/usr/sbin/smtpctl" 

    # ── Dashboard ──────────────────────────────────────────────────────────
    Add-PodeRoute -Method Get -Path "/" -ScriptBlock {
        try {
            # Safely check query params
            $out = ""
            if ($WebEvent.Query.ContainsKey('out')) {
                $out = [uri]::UnescapeDataString($WebEvent.Query['out'])
            }

            # Safely check stats
            $stats = "Loading..."
            try {
                if (Test-Path $SmtpPath) {
                    $stats = & $SmtpPath show stats 2>&1 | Out-String
                } else {
                    $stats = "Error: Path not found at $SmtpPath"
                }
            } catch {
                $stats = "Exception fetching stats: $($_.Exception.Message)"
            }

            # Build HTML with zero nested complex objects
            $html = @"
<!DOCTYPE html>
<html>
<head>
    <title>OpenSMTPD Manager</title>
    <style>
        body { background: #0d1117; color: #c9d1d9; font-family: sans-serif; padding: 20px; }
        .card { background: #161b22; border: 1px solid #30363d; border-radius: 6px; padding: 15px; margin-bottom: 15px; }
        pre { background: #000; padding: 10px; border-radius: 4px; border: 1px solid #444; color: #8b949e; white-space: pre-wrap; }
        button { background: #21262d; color: #58a6ff; border: 1px solid #30363d; padding: 8px 15px; cursor: pointer; border-radius: 4px; }
    </style>
</head>
<body>
    <div style="max-width:800px; margin:auto;">
        <h1>OpenSMTPD Manager</h1>
        <div class="card">
            <h3>Actions</h3>
            <form action="/action" method="post">
                <button name="cmd" value="start">Start</button>
                <button name="cmd" value="stop">Stop</button>
                <button name="cmd" value="reload">Reload</button>
            </form>
        </div>
        <div class="card">
            <h3>System Stats</h3>
            <pre>$($stats)</pre>
        </div>
        <div class="card">
            <h3>Last Command Result</h3>
            <pre style="color:#58a6ff;">$($out)</pre>
        </div>
    </div>
</body>
</html>
"@
            Write-PodeWebResponse -Value $html -ContentType 'text/html'

        } catch {
            # This turns the 500 Error into a visible error message
            $errorMessage = "ROUTE CRASHED: $($_.Exception.Message) `n`n StackTrace: $($_.ScriptStackTrace)"
            Write-PodeWebResponse -Value "<pre>$errorMessage</pre>" -StatusCode 500
        }
    }

    # ── Action Handler ───────────────────────────────────────────────────────
    Add-PodeRoute -Method Post -Path "/action" -ScriptBlock {
        try {
            $action = $WebEvent.Data['cmd']
            if (-not $action) { 
                Move-PodeResponse -Url "/?out=No+command"
                return 
            }

            $res = & $SmtpPath $action 2>&1 | Out-String
            $escaped = [uri]::EscapeDataString($res.Trim())
            Move-PodeResponse -Url "/?out=$escaped"
        } catch {
            $err = [uri]::EscapeDataString("Action Crash: $($_.Exception.Message)")
            Move-PodeResponse -Url "/?out=$err"
        }
    }
}
