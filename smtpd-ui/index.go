package main

import "net/http"

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func loginHTML(errMsg string) string {
	errBlock := ""
	if errMsg != "" {
		errBlock = `<div class="login-err">` + errMsg + `</div>`
	}
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>smtpd manager — login</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;700&family=Syne:wght@400;700;800&display=swap" rel="stylesheet">
<style>
:root {
  --bg:#0a0b0d;--bg1:#0f1115;--bg2:#161820;--bg3:#1e2028;
  --border:#2a2d38;--border2:#363a48;
  --text:#c9cdd8;--text2:#7a8099;--text3:#4a5068;
  --cyan:#00d4c8;--green:#39d98a;--red:#ff5263;
  --font-mono:'JetBrains Mono',monospace;--font-ui:'Syne',sans-serif;
}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{
  background:var(--bg);color:var(--text);font-family:var(--font-mono);
  min-height:100vh;display:flex;align-items:center;justify-content:center;
  overflow:hidden;
}
body::before{
  content:'';position:fixed;inset:0;
  background:repeating-linear-gradient(0deg,transparent,transparent 2px,rgba(0,0,0,.06) 2px,rgba(0,0,0,.06) 4px);
  pointer-events:none;z-index:9999;
}
/* Grid bg */
body::after{
  content:'';position:fixed;inset:0;
  background-image:
    linear-gradient(rgba(0,212,200,.03) 1px,transparent 1px),
    linear-gradient(90deg,rgba(0,212,200,.03) 1px,transparent 1px);
  background-size:40px 40px;
  pointer-events:none;
}
.login-wrap{
  width:360px;
  background:var(--bg1);
  border:1px solid var(--border);
  padding:40px;
  position:relative;
  z-index:1;
  animation:fadeUp .4s ease both;
}
@keyframes fadeUp{
  from{opacity:0;transform:translateY(16px)}
  to{opacity:1;transform:translateY(0)}
}
.login-wrap::before{
  content:'';position:absolute;top:0;left:0;right:0;height:2px;
  background:linear-gradient(90deg,transparent,var(--cyan),transparent);
}
.login-logo{
  font-family:var(--font-ui);font-size:18px;font-weight:800;
  color:var(--cyan);margin-bottom:6px;display:flex;align-items:center;gap:10px;
}
.login-logo-box{
  width:28px;height:28px;border:2px solid var(--cyan);
  display:grid;place-items:center;font-size:12px;position:relative;
}
.login-logo-box::after{
  content:'';position:absolute;inset:3px;background:var(--cyan);opacity:.15;
}
.login-sub{
  font-size:11px;color:var(--text3);margin-bottom:32px;letter-spacing:.04em;
}
.login-label{
  font-size:9px;font-weight:700;letter-spacing:.12em;text-transform:uppercase;
  color:var(--text3);display:block;margin-bottom:6px;
}
.login-input{
  width:100%;background:var(--bg2);border:1px solid var(--border);
  color:var(--text);font-family:var(--font-mono);font-size:13px;
  padding:10px 12px;outline:none;margin-bottom:16px;
  transition:border-color .15s;
}
.login-input:focus{border-color:var(--cyan);}
.login-input::placeholder{color:var(--text3);}
.login-btn{
  width:100%;background:var(--cyan);border:none;color:#0a0b0d;
  font-family:var(--font-mono);font-size:12px;font-weight:700;
  letter-spacing:.08em;text-transform:uppercase;padding:12px;
  cursor:pointer;transition:opacity .15s;margin-top:4px;
}
.login-btn:hover{opacity:.85;}
.login-err{
  background:rgba(255,82,99,.1);border:1px solid rgba(255,82,99,.3);
  color:var(--red);font-size:11px;padding:10px 12px;margin-bottom:20px;
  letter-spacing:.02em;
}
.login-footer{
  text-align:center;font-size:10px;color:var(--text3);
  margin-top:24px;letter-spacing:.04em;
}
</style>
</head>
<body>
<div class="login-wrap">
  <div class="login-logo">
    <div class="login-logo-box">✉</div>
    smtpd<span style="color:var(--text2);font-weight:400">-manage</span>
  </div>
  <div class="login-sub">// OpenSMTPD management console</div>
  ` + errBlock + `
  <form method="POST" action="/login">
    <label class="login-label" for="username">Username</label>
    <input class="login-input" id="username" name="username" type="text"
           autocomplete="username" placeholder="admin" required>
    <label class="login-label" for="password">Password</label>
    <input class="login-input" id="password" name="password" type="password"
           autocomplete="current-password" placeholder="••••••••" required>
    <button class="login-btn" type="submit">→ Sign in</button>
  </form>
  <div class="login-footer">session expires after 8 hours</div>
</div>
</body>
</html>`
}

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>smtpd manager</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;700&family=Syne:wght@400;700;800&display=swap" rel="stylesheet">
<style>
:root {
  --bg:       #0a0b0d;
  --bg1:      #0f1115;
  --bg2:      #161820;
  --bg3:      #1e2028;
  --border:   #2a2d38;
  --border2:  #363a48;
  --text:     #c9cdd8;
  --text2:    #7a8099;
  --text3:    #4a5068;
  --cyan:     #00d4c8;
  --cyan2:    #00a89e;
  --green:    #39d98a;
  --red:      #ff5263;
  --yellow:   #ffd166;
  --purple:   #9d8cff;
  --font-mono: 'JetBrains Mono', monospace;
  --font-ui:   'Syne', sans-serif;
}

*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 13px;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow-x: hidden;
}

body::before {
  content: '';
  position: fixed; inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent,
    transparent 2px,
    rgba(0,0,0,.06) 2px,
    rgba(0,0,0,.06) 4px
  );
  pointer-events: none;
  z-index: 9999;
}

/* ── Header ── */
header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 52px;
  background: var(--bg1);
  border-bottom: 1px solid var(--border);
  position: sticky; top: 0; z-index: 100;
  flex-shrink: 0;
}

.logo {
  font-family: var(--font-ui);
  font-weight: 800;
  font-size: 16px;
  color: var(--cyan);
  letter-spacing: .02em;
  display: flex; align-items: center; gap: 10px;
}
.logo-icon {
  width: 28px; height: 28px;
  border: 2px solid var(--cyan);
  display: grid; place-items: center;
  font-size: 12px; color: var(--cyan);
  position: relative;
}
.logo-icon::after {
  content: '';
  position: absolute;
  inset: 3px;
  background: var(--cyan);
  opacity: .15;
}

.header-meta {
  display: flex; align-items: center; gap: 20px;
  font-size: 11px; color: var(--text2);
}
.status-badge {
  display: flex; align-items: center; gap: 6px;
  padding: 4px 10px;
  border: 1px solid var(--border2);
  background: var(--bg2);
}
.status-dot {
  width: 7px; height: 7px;
  border-radius: 50%;
  background: var(--text3);
  transition: background .3s;
  position: relative;
}
.status-dot.running {
  background: var(--green);
  box-shadow: 0 0 6px var(--green);
}
.status-dot.running::after {
  content: '';
  position: absolute;
  inset: -3px;
  border-radius: 50%;
  border: 1px solid var(--green);
  opacity: .4;
  animation: pulse 2s infinite;
}
@keyframes pulse {
  0%   { transform: scale(1); opacity: .4; }
  50%  { transform: scale(1.6); opacity: 0; }
  100% { transform: scale(1); opacity: 0; }
}

.logout-btn {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: .06em;
  text-transform: uppercase;
  color: var(--text3);
  background: transparent;
  border: 1px solid var(--border);
  padding: 4px 10px;
  cursor: pointer;
  text-decoration: none;
  transition: all .15s;
  display: inline-block;
}
.logout-btn:hover { color: var(--red); border-color: var(--red); }

#clock { color: var(--text3); font-size: 11px; }

/* ── Layout ── */
.app { display: flex; flex: 1; overflow: hidden; }

nav {
  width: 220px;
  flex-shrink: 0;
  background: var(--bg1);
  border-right: 1px solid var(--border);
  padding: 20px 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}
.nav-section { margin-bottom: 4px; }
.nav-label {
  font-size: 9px;
  font-weight: 700;
  letter-spacing: .12em;
  color: var(--text3);
  text-transform: uppercase;
  padding: 12px 20px 6px;
}
.nav-item {
  display: flex; align-items: center; gap: 10px;
  padding: 9px 20px;
  cursor: pointer;
  color: var(--text2);
  font-size: 12px;
  transition: all .15s;
  border-left: 2px solid transparent;
  user-select: none;
  white-space: nowrap;
}
.nav-item:hover { color: var(--text); background: var(--bg2); }
.nav-item.active {
  color: var(--cyan);
  background: rgba(0,212,200,.06);
  border-left-color: var(--cyan);
}
.nav-item .ni-icon { width: 14px; text-align: center; opacity: .7; flex-shrink: 0; }

/* ── Main panel ── */
main { flex: 1; display: flex; flex-direction: column; overflow: hidden; }

.panel { flex: 1; overflow-y: auto; padding: 28px; display: none; }
.panel.active { display: block; }

.panel-title {
  font-family: var(--font-ui);
  font-size: 20px; font-weight: 700; color: var(--text);
  margin-bottom: 4px; letter-spacing: -.01em;
}
.panel-desc { color: var(--text3); font-size: 11px; margin-bottom: 24px; letter-spacing: .02em; }

/* ── Action grid ── */
.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  margin-bottom: 24px;
}
.action-btn {
  background: var(--bg2);
  border: 1px solid var(--border);
  color: var(--text);
  padding: 14px 16px;
  cursor: pointer;
  text-align: left;
  font-family: var(--font-mono);
  font-size: 12px;
  transition: all .15s;
  display: flex; flex-direction: column; gap: 6px;
  position: relative; overflow: hidden;
}
.action-btn::before {
  content: '';
  position: absolute; top: 0; left: 0;
  width: 3px; height: 100%;
  background: var(--cyan);
  opacity: 0; transition: opacity .15s;
}
.action-btn:hover { border-color: var(--border2); background: var(--bg3); color: var(--cyan); }
.action-btn:hover::before { opacity: 1; }
.action-btn:active { transform: scale(.99); }
.action-btn.danger:hover { color: var(--red); }
.action-btn.danger::before { background: var(--red); }
.action-btn.warn:hover { color: var(--yellow); }
.action-btn.warn::before { background: var(--yellow); }
.action-btn.ok:hover { color: var(--green); }
.action-btn.ok::before { background: var(--green); }

.btn-label { font-weight: 700; font-size: 11px; letter-spacing: .06em; text-transform: uppercase; }
.btn-desc { font-size: 10px; color: var(--text3); line-height: 1.4; }

/* ── ID input group ── */
.id-group { display: flex; gap: 8px; align-items: flex-end; margin-bottom: 12px; flex-wrap: wrap; }
.id-group label { font-size: 10px; color: var(--text3); display: block; margin-bottom: 4px; letter-spacing: .06em; text-transform: uppercase; }
.id-input {
  background: var(--bg2); border: 1px solid var(--border);
  color: var(--text); font-family: var(--font-mono); font-size: 12px;
  padding: 9px 12px; outline: none; width: 220px; transition: border-color .15s;
}
.id-input:focus { border-color: var(--cyan); }
.id-input::placeholder { color: var(--text3); }
.id-btn {
  background: var(--bg3); border: 1px solid var(--border2);
  color: var(--text); font-family: var(--font-mono); font-size: 11px;
  padding: 9px 14px; cursor: pointer; text-transform: uppercase;
  letter-spacing: .06em; transition: all .15s; white-space: nowrap;
}
.id-btn:hover { border-color: var(--cyan); color: var(--cyan); }
.id-btn.danger:hover { border-color: var(--red); color: var(--red); }
.id-btn.warn:hover { border-color: var(--yellow); color: var(--yellow); }

/* ── Terminal output ── */
.terminal {
  background: var(--bg1); border: 1px solid var(--border);
  padding: 16px; font-family: var(--font-mono); font-size: 12px;
  line-height: 1.7; color: var(--green); white-space: pre-wrap;
  word-break: break-all; min-height: 160px; max-height: 400px;
  overflow-y: auto; position: relative;
}
.terminal.error { color: var(--red); }
.terminal-header {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 8px;
}
.terminal-label { font-size: 9px; font-weight: 700; letter-spacing: .12em; text-transform: uppercase; color: var(--text3); }
.terminal-clear {
  font-size: 9px; color: var(--text3); cursor: pointer;
  letter-spacing: .06em; text-transform: uppercase;
  padding: 2px 6px; border: 1px solid transparent;
}
.terminal-clear:hover { border-color: var(--border2); color: var(--text2); }
.terminal::before { content: '$ '; color: var(--cyan); opacity: .5; }

/* ── Log viewer (full-height, different layout) ── */
#panel-logs {
  padding: 0;
  display: none;
  flex-direction: column;
  height: 100%;
}
#panel-logs.active { display: flex; }

.log-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  padding: 14px 24px;
  background: var(--bg1);
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  gap: 16px;
  flex-wrap: wrap;
}
.log-title {
  font-family: var(--font-ui);
  font-size: 16px; font-weight: 700; color: var(--text);
}
.log-controls { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.log-path { font-size: 10px; color: var(--text3); letter-spacing: .04em; }
.log-badge {
  font-size: 9px; font-weight: 700; letter-spacing: .1em; text-transform: uppercase;
  padding: 3px 8px; border: 1px solid var(--border2);
  display: flex; align-items: center; gap: 5px;
}
.log-badge.live { border-color: rgba(57,217,138,.4); color: var(--green); }
.log-badge.live .lb-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: var(--green);
  animation: pulse 1.5s infinite;
}
.log-badge.paused { border-color: var(--border2); color: var(--text3); }

.log-filter {
  background: var(--bg2); border: 1px solid var(--border);
  color: var(--text); font-family: var(--font-mono); font-size: 11px;
  padding: 6px 10px; outline: none; width: 200px;
  transition: border-color .15s;
}
.log-filter:focus { border-color: var(--cyan); }
.log-filter::placeholder { color: var(--text3); }

.log-ctrl-btn {
  font-family: var(--font-mono); font-size: 10px; letter-spacing: .06em;
  text-transform: uppercase; padding: 6px 12px;
  background: var(--bg3); border: 1px solid var(--border2);
  color: var(--text2); cursor: pointer; transition: all .15s;
}
.log-ctrl-btn:hover { color: var(--text); border-color: var(--border2); }
.log-ctrl-btn.active-btn { color: var(--cyan); border-color: var(--cyan); }

.log-stream {
  flex: 1;
  overflow-y: auto;
  padding: 0;
  background: var(--bg);
  font-family: var(--font-mono);
  font-size: 11.5px;
  line-height: 1.65;
}
.log-line {
  display: flex;
  padding: 1px 20px;
  border-bottom: 1px solid transparent;
  transition: background .1s;
  white-space: pre-wrap;
  word-break: break-all;
}
.log-line:hover { background: var(--bg2); }
.log-line.hidden { display: none; }

/* Log coloring */
.log-line.ll-err  { color: var(--red); }
.log-line.ll-warn { color: var(--yellow); }
.log-line.ll-ok   { color: var(--green); }
.log-line.ll-info { color: var(--text2); }

.log-ln { color: var(--text3); user-select: none; min-width: 48px; margin-right: 16px; text-align: right; flex-shrink: 0; }

.log-empty {
  padding: 40px 20px;
  color: var(--text3);
  font-size: 12px;
  text-align: center;
}

/* ── Spinner ── */
.spinner { display: none; align-items: center; gap: 8px; font-size: 11px; color: var(--text2); margin-bottom: 10px; }
.spinner.active { display: flex; }
.spin-ring {
  width: 14px; height: 14px;
  border: 2px solid var(--border);
  border-top-color: var(--cyan);
  border-radius: 50%;
  animation: spin .6s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── Toast ── */
#toast {
  position: fixed; bottom: 24px; right: 24px;
  background: var(--bg2); border: 1px solid var(--border2);
  padding: 12px 18px; font-size: 12px; max-width: 360px;
  opacity: 0; transform: translateY(8px);
  transition: all .25s; pointer-events: none; z-index: 1000;
  border-left: 3px solid var(--cyan);
}
#toast.show { opacity: 1; transform: translateY(0); }
#toast.ok  { border-left-color: var(--green); }
#toast.err { border-left-color: var(--red); }

/* ── Misc ── */
.divider { border: none; border-top: 1px solid var(--border); margin: 20px 0; }
.section-hd {
  font-size: 9px; font-weight: 700; letter-spacing: .12em;
  text-transform: uppercase; color: var(--text3);
  margin-bottom: 10px; margin-top: 20px;
  display: flex; align-items: center; gap: 8px;
}
.section-hd::after { content: ''; flex: 1; height: 1px; background: var(--border); }

::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border2); }
::-webkit-scrollbar-thumb:hover { background: var(--text3); }
</style>
</head>
<body>

<header>
  <div class="logo">
    <div class="logo-icon">✉</div>
    smtpd<span style="color:var(--text2);font-weight:400">-manage</span>
  </div>
  <div class="header-meta">
    <div class="status-badge">
      <div class="status-dot" id="statusDot"></div>
      <span id="statusText">checking…</span>
    </div>
    <span id="hostname" style="color:var(--text3)"></span>
    <span id="clock"></span>
    <a class="logout-btn" href="/logout">⎋ logout</a>
  </div>
</header>

<div class="app">
  <!-- Sidebar -->
  <nav>
    <div class="nav-section">
      <div class="nav-label">Control</div>
      <div class="nav-item active" onclick="showPanel('daemon',this)">
        <span class="ni-icon">⬡</span> Daemon
      </div>
      <div class="nav-item" onclick="showPanel('queue',this)">
        <span class="ni-icon">≡</span> Queue
      </div>
    </div>
    <div class="nav-section">
      <div class="nav-label">Observe</div>
      <div class="nav-item" onclick="showPanel('monitoring',this)">
        <span class="ni-icon">◈</span> Monitoring
      </div>
      <div class="nav-item" onclick="showPanel('flow',this)">
        <span class="ni-icon">⇄</span> Mail Flow
      </div>
      <div class="nav-item" onclick="showPanel('logs',this)">
        <span class="ni-icon">▤</span> Log Stream
      </div>
    </div>
    <div class="nav-section">
      <div class="nav-label">System</div>
      <div class="nav-item" onclick="showPanel('config',this)">
        <span class="ni-icon">⚙</span> Config
      </div>
    </div>
  </nav>

  <main>

    <!-- Daemon -->
    <section class="panel active" id="panel-daemon">
      <div class="panel-title">Daemon Control</div>
      <div class="panel-desc">// manage the smtpd process — root privileges required</div>
      <div class="action-grid">
        <button class="action-btn ok" onclick="run('start')">
          <span class="btn-label">▶ Start</span>
          <span class="btn-desc">Launch the smtpd daemon</span>
        </button>
        <button class="action-btn danger" onclick="run('stop')">
          <span class="btn-label">■ Stop</span>
          <span class="btn-desc">Gracefully halt smtpd; queue preserved</span>
        </button>
        <button class="action-btn warn" onclick="run('restart')">
          <span class="btn-label">↺ Restart</span>
          <span class="btn-desc">Stop then start; drops active connections</span>
        </button>
        <button class="action-btn" onclick="run('reload')">
          <span class="btn-label">⟳ Reload</span>
          <span class="btn-desc">Re-read smtpd.conf without connection drop</span>
        </button>
      </div>
      <div class="terminal-header">
        <span class="terminal-label">output</span>
        <span class="terminal-clear" onclick="clearTerm('daemon')">clear</span>
      </div>
      <div class="spinner" id="spin-daemon"><div class="spin-ring"></div><span>executing…</span></div>
      <div class="terminal" id="term-daemon">(idle — run a command above)</div>
    </section>

    <!-- Queue -->
    <section class="panel" id="panel-queue">
      <div class="panel-title">Queue Management</div>
      <div class="panel-desc">// inspect, flush, bounce, and remove mail queue entries</div>
      <div class="action-grid">
        <button class="action-btn" onclick="run('show-queue')">
          <span class="btn-label">⊞ List Queue</span>
          <span class="btn-desc">Verbose envelope listing</span>
        </button>
        <button class="action-btn ok" onclick="run('queue-flush')">
          <span class="btn-label">⚡ Flush All</span>
          <span class="btn-desc">Schedule immediate delivery for all messages</span>
        </button>
        <button class="action-btn warn" onclick="run('queue-hold')">
          <span class="btn-label">⏸ Hold MTA</span>
          <span class="btn-desc">Pause outbound delivery; messages accumulate</span>
        </button>
        <button class="action-btn ok" onclick="run('queue-resume')">
          <span class="btn-label">▶ Resume MTA</span>
          <span class="btn-desc">Resume paused outbound delivery</span>
        </button>
      </div>
      <div class="section-hd">Message operations</div>
      <div class="id-group">
        <div>
          <label>Queue Message ID</label>
          <input class="id-input" id="queueId" placeholder="e.g. abc123def456" spellcheck="false">
        </div>
        <button class="id-btn" onclick="runWithId('schedule','queueId')">⚡ Schedule</button>
        <button class="id-btn warn" onclick="runWithId('bounce','queueId')">✕ Bounce</button>
        <button class="id-btn danger" onclick="runWithId('remove','queueId')">⊗ Remove</button>
      </div>
      <div class="terminal-header">
        <span class="terminal-label">output</span>
        <span class="terminal-clear" onclick="clearTerm('queue')">clear</span>
      </div>
      <div class="spinner" id="spin-queue"><div class="spin-ring"></div><span>executing…</span></div>
      <div class="terminal" id="term-queue">(idle — run a command above)</div>
    </section>

    <!-- Monitoring -->
    <section class="panel" id="panel-monitoring">
      <div class="panel-title">Monitoring</div>
      <div class="panel-desc">// runtime status, statistics, and log verbosity</div>
      <div class="action-grid">
        <button class="action-btn" onclick="run('status')">
          <span class="btn-label">◈ Status</span>
          <span class="btn-desc">Daemon state and queue overview</span>
        </button>
        <button class="action-btn" onclick="run('show-stats')">
          <span class="btn-label">◫ Full Stats</span>
          <span class="btn-desc">All runtime counters from smtpctl</span>
        </button>
        <button class="action-btn" onclick="run('show-queue')">
          <span class="btn-label">≡ Queue</span>
          <span class="btn-desc">Verbose envelope-level listing</span>
        </button>
      </div>
      <div class="section-hd">Log verbosity</div>
      <div class="action-grid">
        <button class="action-btn" onclick="run('log-brief')">
          <span class="btn-label">○ Brief</span>
          <span class="btn-desc">Minimal — default production level</span>
        </button>
        <button class="action-btn warn" onclick="run('log-verbose')">
          <span class="btn-label">◎ Verbose</span>
          <span class="btn-desc">Connection and delivery steps</span>
        </button>
        <button class="action-btn danger" onclick="run('log-trace')">
          <span class="btn-label">● Trace</span>
          <span class="btn-desc">Full trace — very high volume output</span>
        </button>
      </div>
      <div class="terminal-header">
        <span class="terminal-label">output</span>
        <span class="terminal-clear" onclick="clearTerm('monitoring')">clear</span>
      </div>
      <div class="spinner" id="spin-monitoring"><div class="spin-ring"></div><span>executing…</span></div>
      <div class="terminal" id="term-monitoring">(idle — run a command above)</div>
    </section>

    <!-- Flow -->
    <section class="panel" id="panel-flow">
      <div class="panel-title">Mail Flow Control</div>
      <div class="panel-desc">// pause or resume inbound SMTP and outbound MTA</div>
      <div class="section-hd">SMTP listener — inbound</div>
      <div class="action-grid">
        <button class="action-btn warn" onclick="run('pause-smtp')">
          <span class="btn-label">⏸ Pause SMTP</span>
          <span class="btn-desc">Stop accepting new inbound connections</span>
        </button>
        <button class="action-btn ok" onclick="run('resume-smtp')">
          <span class="btn-label">▶ Resume SMTP</span>
          <span class="btn-desc">Accept inbound connections again</span>
        </button>
      </div>
      <div class="section-hd">MTA delivery — outbound</div>
      <div class="action-grid">
        <button class="action-btn warn" onclick="run('pause-mta')">
          <span class="btn-label">⏸ Pause MTA</span>
          <span class="btn-desc">Stop outbound delivery; messages queue up</span>
        </button>
        <button class="action-btn ok" onclick="run('resume-mta')">
          <span class="btn-label">▶ Resume MTA</span>
          <span class="btn-desc">Resume outbound delivery</span>
        </button>
      </div>
      <div class="terminal-header">
        <span class="terminal-label">output</span>
        <span class="terminal-clear" onclick="clearTerm('flow')">clear</span>
      </div>
      <div class="spinner" id="spin-flow"><div class="spin-ring"></div><span>executing…</span></div>
      <div class="terminal" id="term-flow">(idle — run a command above)</div>
    </section>

    <!-- ── Log Stream ── -->
    <section class="panel" id="panel-logs">
      <div class="log-toolbar">
        <div>
          <div class="log-title">Log Stream</div>
          <div class="log-path">▤ /var/log/opensmtpd/opensmtpd.log</div>
        </div>
        <div class="log-controls">
          <input class="log-filter" id="logFilter" placeholder="filter lines…" oninput="applyFilter()" spellcheck="false">
          <button class="log-ctrl-btn" id="btnPause" onclick="togglePause()">⏸ Pause</button>
          <button class="log-ctrl-btn" onclick="clearLog()">⊗ Clear</button>
          <button class="log-ctrl-btn" onclick="scrollToBottom()">↓ Bottom</button>
          <div class="log-badge live" id="logBadge">
            <div class="lb-dot"></div>
            <span id="logBadgeText">live</span>
          </div>
        </div>
      </div>
      <div class="log-stream" id="logStream">
        <div class="log-empty" id="logEmpty">Connecting to log stream…</div>
      </div>
    </section>

    <!-- Config -->
    <section class="panel" id="panel-config">
      <div class="panel-title">Config &amp; Tools</div>
      <div class="panel-desc">// syntax checking and configuration inspection</div>
      <div class="action-grid">
        <button class="action-btn" onclick="run('check-config')">
          <span class="btn-label">✓ Check Config</span>
          <span class="btn-desc">Syntax-check /etc/mail/smtpd.conf (root)</span>
        </button>
        <button class="action-btn" onclick="run('show-config')">
          <span class="btn-label">⊡ Show Config</span>
          <span class="btn-desc">Print effective running configuration (root)</span>
        </button>
      </div>
      <div class="terminal-header">
        <span class="terminal-label">output</span>
        <span class="terminal-clear" onclick="clearTerm('config')">clear</span>
      </div>
      <div class="spinner" id="spin-config"><div class="spin-ring"></div><span>executing…</span></div>
      <div class="terminal" id="term-config">(idle — run a command above)</div>
    </section>

  </main>
</div>

<div id="toast"></div>

<script>
// ── Panel switching ───────────────────────────────────────────────────────────
let currentPanel = 'daemon';
let logStarted = false;

function showPanel(name, el) {
  document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
  document.getElementById('panel-' + name).classList.add('active');
  el.classList.add('active');
  currentPanel = name;
  if (name === 'logs' && !logStarted) {
    startLogStream();
    logStarted = true;
  }
}

// ── Clock ─────────────────────────────────────────────────────────────────────
function tick() {
  const now = new Date();
  document.getElementById('clock').textContent =
    now.getFullYear() + '-' +
    String(now.getMonth()+1).padStart(2,'0') + '-' +
    String(now.getDate()).padStart(2,'0') + ' ' +
    String(now.getHours()).padStart(2,'0') + ':' +
    String(now.getMinutes()).padStart(2,'0') + ':' +
    String(now.getSeconds()).padStart(2,'0');
}
tick(); setInterval(tick, 1000);

// ── Status polling ────────────────────────────────────────────────────────────
async function pollStatus() {
  try {
    const r = await fetch('/api/status');
    if (r.status === 401) { location.href = '/login'; return; }
    const d = await r.json();
    document.getElementById('statusDot').className = 'status-dot' + (d.running ? ' running' : '');
    document.getElementById('statusText').textContent = d.running ? 'running' : 'stopped';
    document.getElementById('hostname').textContent = d.hostname || '';
  } catch {}
}
pollStatus(); setInterval(pollStatus, 5000);

// ── Toast ─────────────────────────────────────────────────────────────────────
let toastTimer;
function showToast(msg, type='ok') {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'show ' + type;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.className = ''; }, 3000);
}

// ── Terminal helpers ──────────────────────────────────────────────────────────
function setTerm(panel, text, isError) {
  const el = document.getElementById('term-' + panel);
  el.textContent = text;
  el.className = 'terminal' + (isError ? ' error' : '');
}
function clearTerm(panel) { setTerm(panel, '(cleared)', false); }
function setSpin(panel, active) {
  document.getElementById('spin-' + panel).className = 'spinner' + (active ? ' active' : '');
}

// ── Action runner ─────────────────────────────────────────────────────────────
async function run(action, id, panel) {
  if (!panel) panel = currentPanel;
  setSpin(panel, true);
  setTerm(panel, 'running: ' + action + (id ? ' ' + id : '') + '…', false);
  try {
    const res = await fetch('/api/action', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action, id: id || '' })
    });
    if (res.status === 401) { location.href = '/login'; return; }
    const data = await res.json();
    setSpin(panel, false);
    setTerm(panel, data.output, !data.success);
    showToast(data.success ? '✓ ' + action : '✗ ' + action + ' failed', data.success ? 'ok' : 'err');
    pollStatus();
  } catch (e) {
    setSpin(panel, false);
    setTerm(panel, 'Network error: ' + e.message, true);
    showToast('✗ Network error', 'err');
  }
}
function runWithId(action, inputId) {
  const id = document.getElementById(inputId).value.trim();
  if (!id) { showToast('✗ Enter a message ID first', 'err'); return; }
  run(action, id, currentPanel);
}

// ── Log stream ────────────────────────────────────────────────────────────────
let logPaused = false;
let logLineCount = 0;
const MAX_LOG_LINES = 2000;
let pendingLines = [];   // buffer while paused
let filterText = '';
let autoScroll = true;

const logStream = document.getElementById('logStream');
const logEmpty  = document.getElementById('logEmpty');

function classifyLine(text) {
  const t = text.toLowerCase();
  if (/error|fatal|fail|reject|denied/.test(t))  return 'll-err';
  if (/warn|timeout|slow|defer/.test(t))          return 'll-warn';
  if (/delivered|sent|accept|connect|auth ok/.test(t)) return 'll-ok';
  return 'll-info';
}

function addLogLine(text) {
  logEmpty.style.display = 'none';
  logLineCount++;

  // Prune old lines if over limit
  const lines = logStream.querySelectorAll('.log-line');
  if (lines.length >= MAX_LOG_LINES) {
    lines[0].remove();
  }

  const div = document.createElement('div');
  div.className = 'log-line ' + classifyLine(text);

  const ln = document.createElement('span');
  ln.className = 'log-ln';
  ln.textContent = logLineCount;

  const txt = document.createElement('span');
  txt.textContent = text;

  div.appendChild(ln);
  div.appendChild(txt);

  // Apply filter
  if (filterText && !text.toLowerCase().includes(filterText)) {
    div.classList.add('hidden');
  }

  logStream.appendChild(div);

  if (autoScroll && !logPaused) {
    logStream.scrollTop = logStream.scrollHeight;
  }
}

function applyFilter() {
  filterText = document.getElementById('logFilter').value.toLowerCase().trim();
  logStream.querySelectorAll('.log-line').forEach(el => {
    const txt = el.querySelector('span:last-child').textContent.toLowerCase();
    el.classList.toggle('hidden', filterText !== '' && !txt.includes(filterText));
  });
}

function clearLog() {
  logStream.querySelectorAll('.log-line').forEach(el => el.remove());
  logEmpty.style.display = '';
  logEmpty.textContent = '(cleared)';
  logLineCount = 0;
}

function scrollToBottom() {
  logStream.scrollTop = logStream.scrollHeight;
}

function togglePause() {
  logPaused = !logPaused;
  const btn = document.getElementById('btnPause');
  const badge = document.getElementById('logBadge');
  const badgeTxt = document.getElementById('logBadgeText');

  if (logPaused) {
    btn.textContent = '▶ Resume';
    btn.classList.add('active-btn');
    badge.className = 'log-badge paused';
    badgeTxt.textContent = 'paused';
  } else {
    btn.textContent = '⏸ Pause';
    btn.classList.remove('active-btn');
    badge.className = 'log-badge live';
    badgeTxt.textContent = 'live';
    // Flush buffered lines
    pendingLines.forEach(l => addLogLine(l));
    pendingLines = [];
    scrollToBottom();
  }
}

// Track auto-scroll: if user scrolls up, stop auto-scrolling
logStream.addEventListener('scroll', () => {
  const threshold = 60;
  autoScroll = logStream.scrollHeight - logStream.scrollTop - logStream.clientHeight < threshold;
});

let evtSource = null;
let reconnectDelay = 1000;

function startLogStream() {
  if (evtSource) { evtSource.close(); }

  logEmpty.style.display = '';
  logEmpty.textContent = 'Connecting…';

  evtSource = new EventSource('/api/logs');

  evtSource.onopen = () => {
    reconnectDelay = 1000;
    document.getElementById('logBadge').className = 'log-badge live';
    document.getElementById('logBadgeText').textContent = 'live';
  };

  evtSource.onmessage = (e) => {
    const line = e.data;
    if (logPaused) {
      pendingLines.push(line);
      document.getElementById('logBadgeText').textContent = 'paused +' + pendingLines.length;
    } else {
      addLogLine(line);
    }
  };

  evtSource.onerror = () => {
    document.getElementById('logBadge').className = 'log-badge paused';
    document.getElementById('logBadgeText').textContent = 'reconnecting…';
    evtSource.close();
    setTimeout(() => {
      if (document.getElementById('panel-logs').classList.contains('active') || logStarted) {
        reconnectDelay = Math.min(reconnectDelay * 2, 16000);
        startLogStream();
      }
    }, reconnectDelay);
  };
}
</script>
</body>
</html>`
