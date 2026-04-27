#!/usr/bin/pwsh

# --- Configuration ---
$Port = 8080
$LogPath = "/var/log/opensmptd/opensmtpd.log" # Common location; adjust if your syslog differs
$Hostname = hostname

# --- Initialization ---
$Listener = New-Object System.Net.HttpListener
$Listener.Prefixes.Add("http://*:$Port/")

try {
    $Listener.Start()
} catch {
    Write-Error "Failed to start listener. Are you root? Is port $Port busy?"
    Exit
}

Write-Host "Started Maillog Streamer" -ForegroundColor Cyan
Write-Host "Listening on: http://<your_ip>:$Port/" -ForegroundColor Yellow
Write-Host "Streaming: $LogPath" -ForegroundColor Green
Write-Host "Press Ctrl+C to stop."

# --- Helper to handle network write errors safely ---
function Send-SSEEvent {
    param($Writer, $Data)
    try {
        # SSE Format: "data: <content>\n\n"
        $Writer.WriteLine("data: $Data")
        $Writer.WriteLine() 
        $Writer.Flush()
        return $true
    } catch {
        # Connection likely lost
        return $false
    }
}

# --- Main Loop ---
try {
    while ($Listener.IsListening) {
        $Context = $Listener.GetContext()
        $Request = $Context.Request
        $Response = $Context.Response

        # --- Route 1: The UI (HTML/JS) ---
        if ($Request.Url.LocalPath -eq "/") {
            $Response.ContentType = "text/html"
            
            $Html = @"
<!DOCTYPE html>
<html lang="en">
<head>
    <title>$Hostname :: Live Maillog</title>
    <style>
        :root {
            --bg-color: #0d1117;
            --text-color: #c9d1d9;
            --terminal-green: #26fb13;
            --border-color: #30363d;
            --timestamp-color: #8b949e;
        }

        body {
            background-color: var(--bg-color);
            color: var(--text-color);
            font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
            margin: 0;
            padding: 0;
            height: 100vh;
            display: flex;
            flex-direction: column;
        }

        header {
            background-color: #161b22;
            padding: 10px 20px;
            border-bottom: 1px solid var(--border-color);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        h1 { margin: 0; font-size: 1.2rem; color: var(--terminal-green); }
        .status { font-size: 0.9rem; color: var(--timestamp-color); }
        .blink { animation: blinker 1s linear infinite; }
        @keyframes blinker { 50% { opacity: 0; } }

        #log-container {
            flex-grow: 1;
            overflow-y: auto;
            padding: 20px;
            scroll-behavior: smooth;
        }

        #log-output {
            white-space: pre-wrap;
            margin: 0;
            word-wrap: break-word;
        }

        /* Basic Maillog Highlighting */
        .line { display: block; margin-bottom: 2px; line-height: 1.4; }
        .line:hover { background-color: #1c2128; }
        
        /* Highlight timestamps (assuming standard syslog format) */
        .ts { color: var(--timestamp-color); }
        
        /* Highlight common SMTP status words */
        .stat-ok { color: var(--terminal-green); font-weight: bold;}
        .stat-err { color: #ff7b72; font-weight: bold;}
        .stat-warn { color: #d29922; font-weight: bold;}

    </style>
</head>
<body>
    <header>
        <h1>SYSTEM MAILLOG :: $Hostname</h1>
        <div class="status">Source: $LogPath | <span class="blink">●</span> LIVE</div>
    </header>
    <div id="log-container">
        <pre id="log-output">Connecting to stream...</pre>
    </div>

    <script>
        const logOutput = document.getElementById('log-output');
        const logContainer = document.getElementById('log-container');
        const eventSource = new EventSource('/stream');

        // Function to process and highlight lines
        function formatLine(rawText) {
            if (!rawText || rawText.trim() === "") return "";
            
            let escaped = rawText
                .replace(/&/g, "&amp;")
                .replace(/</g, "&lt;")
                .replace(/>/g, "&gt;");

            // Highlight timestamp (Jan 01 12:00:00)
            escaped = escaped.replace(/^([A-Z][a-z]{2}\s+\d+\s\d{2}:\d{2}:\d{2})/, '<span class="ts">$1</span>');

            // Semantic highlighting
            escaped = escaped.replace(/\b(ok|status=sent|accepted)\b/gi, '<span class="stat-ok">$1</span>');
            escaped = escaped.replace(/\b(error|failed|rejected|fatal|status=bounced)\b/gi, '<span class="stat-err">$1</span>');
            escaped = escaped.replace(/\b(warning|deferred)\b/gi, '<span class="stat-warn">$1</span>');

            return '<span class="line">' + escaped + '</span>';
        }

        eventSource.onopen = function() {
            logOutput.innerHTML = '<span class="status">--- Connection Established ---</span>\n';
        };

        eventSource.onmessage = function(event) {
            const formatted = formatLine(event.data);
            if (formatted) {
                logOutput.innerHTML += formatted;
                
                // Auto-scroll to bottom
                logContainer.scrollTop = logContainer.scrollHeight;
            }
        };

        eventSource.onerror = function(err) {
            console.error("EventSource failed:", err);
            logOutput.innerHTML += '<span class="stat-err">--- Connection Lost. Reconnecting... ---</span>\n';
        };
    </script>
</body>
</html>
"@
            $Buffer = [System.Text.Encoding]::UTF8.GetBytes($Html)
            $Response.ContentLength64 = $Buffer.Length
            $Response.OutputStream.Write($Buffer, 0, $Buffer.Length)
            $Response.Close()
        }

        # --- Route 2: The Stream (Server-Sent Events) ---
        elseif ($Request.Url.LocalPath -eq "/stream") {
            
            # Critical headers for SSE
            $Response.ContentType = "text/event-stream"
            $Response.Headers.Add("Cache-Control", "no-cache")
            $Response.Headers.Add("Connection", "keep-alive")
            
            # Force the response headers out immediately
            $Response.OutputStream.Flush()
            
            $Writer = New-Object System.IO.StreamWriter($Response.OutputStream)
            
            Write-Host "New stream client connected." -ForegroundColor Gray

            # Use tail -f natively on POSIX
            $TailProcessInfo = New-Object System.Diagnostics.ProcessStartInfo
            $TailProcessInfo.FileName = "tail"
            $TailProcessInfo.Arguments = "-n 20 -f $LogPath" # Start with last 20 lines
            $TailProcessInfo.RedirectStandardOutput = $true
            $TailProcessInfo.UseShellExecute = $false
            
            $TailProcess = [System.Diagnostics.Process]::Start($TailProcessInfo)
            
            # Send initial greeting
            if (-not (Send-SSEEvent -Writer $Writer -Data "--- Tailing $LogPath ---")) {
                 $TailProcess.Kill(); $Response.Close(); continue
            }

            # Stream the output of tail -f to the web response
            while (-not $TailProcess.StandardOutput.EndOfStream) {
                $Line = $TailProcess.StandardOutput.ReadLine()
                
                # If network send fails, client disconnected
                if (-not (Send-SSEEvent -Writer $Writer -Data $Line)) {
                    Write-Host "Stream client disconnected." -ForegroundColor Gray
                    break
                }
            }

            # Clean up when stream ends (client disconnect)
            $TailProcess.Kill()
            $Response.Close()
        }
        else {
            $Response.StatusCode = 404
            $Response.Close()
        }
    }
}
finally {
    # Ensure listener stops on Ctrl+C
    if ($Listener.IsListening) {
        $Listener.Stop()
    }
    Write-Host "Server stopped." -ForegroundColor Red
}
