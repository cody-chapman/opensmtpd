package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ── Config ─────────────────────────────────────────────────────────────────────
// Change these or load from env/flags before deploying.
const (
	adminUser  = "admin"
	adminPass  = "smtpd"
	logPath    = "/var/log/opensmtpd/opensmtpd.log"
	listenAddr = ":8080"
	sessionTTL = 8 * time.Hour
)

// ── Session store ──────────────────────────────────────────────────────────────
type session struct{ created time.Time }

var (
	sessions   = map[string]*session{}
	sessionsMu sync.Mutex
)

func newSession() string {
	b := make([]byte, 24)
	rand.Read(b)
	tok := hex.EncodeToString(b)
	sessionsMu.Lock()
	sessions[tok] = &session{created: time.Now()}
	sessionsMu.Unlock()
	return tok
}

func validSession(tok string) bool {
	if tok == "" {
		return false
	}
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	s, ok := sessions[tok]
	if !ok {
		return false
	}
	if time.Since(s.created) > sessionTTL {
		delete(sessions, tok)
		return false
	}
	return true
}

func deleteSession(tok string) {
	sessionsMu.Lock()
	delete(sessions, tok)
	sessionsMu.Unlock()
}

func sessionToken(r *http.Request) string {
	c, err := r.Cookie("smtpd_session")
	if err != nil {
		return ""
	}
	return c.Value
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validSession(sessionToken(r)) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// ── Login / logout ─────────────────────────────────────────────────────────────
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := r.FormValue("username")
		pass := r.FormValue("password")
		uOK := subtle.ConstantTimeCompare([]byte(user), []byte(adminUser)) == 1
		pOK := subtle.ConstantTimeCompare([]byte(pass), []byte(adminPass)) == 1
		if uOK && pOK {
			tok := newSession()
			http.SetCookie(w, &http.Cookie{
				Name:     "smtpd_session",
				Value:    tok,
				Path:     "/",
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   int(sessionTTL.Seconds()),
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(loginHTML("Invalid username or password.")))
		return
	}
	if validSession(sessionToken(r)) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(loginHTML("")))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	deleteSession(sessionToken(r))
	http.SetCookie(w, &http.Cookie{Name: "smtpd_session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ── Command helpers ────────────────────────────────────────────────────────────
type CommandResult struct {
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
}

func runCmd(name string, args ...string) CommandResult {
	cmd := exec.Command(name, args...)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	err := cmd.Run()
	code := 0
	if err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			code = ex.ExitCode()
		} else {
			code = 1
		}
	}
	combined := strings.TrimSpace(out.String())
	if e := strings.TrimSpace(errb.String()); e != "" {
		if combined != "" {
			combined += "\n" + e
		} else {
			combined = e
		}
	}
	if combined == "" {
		if err == nil {
			combined = "(command completed with no output)"
		} else {
			combined = fmt.Sprintf("error: %v", err)
		}
	}
	return CommandResult{Output: combined, Success: err == nil, Code: code}
}

func smtpctl(args ...string) CommandResult { return runCmd("smtpctl", args...) }
func daemonRunning() bool                  { return exec.Command("smtpctl", "show", "stats").Run() == nil }

// ── SSE log stream ─────────────────────────────────────────────────────────────
func handleLogStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	f, err := os.Open(logPath)
	if err != nil {
		fmt.Fprintf(w, "data: [error] cannot open %s: %v\n\n", logPath, err)
		flusher.Flush()
		return
	}
	defer f.Close()

	// Send last 150 lines as history
	for _, line := range tailLines(f, 150) {
		fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(line, "\n", " "))
	}
	flusher.Flush()

	// Seek to end for live tail
	f.Seek(0, io.SeekEnd)
	scanner := bufio.NewScanner(f)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(line, "\n", " "))
				flusher.Flush()
			}
			scanner = bufio.NewScanner(f)
		}
	}
}

func tailLines(f *os.File, n int) []string {
	f.Seek(0, io.SeekStart)
	sc := bufio.NewScanner(f)
	var all []string
	for sc.Scan() {
		all = append(all, sc.Text())
	}
	if len(all) <= n {
		return all
	}
	return all[len(all)-n:]
}

// ── API action ─────────────────────────────────────────────────────────────────
type ActionRequest struct {
	Action string `json:"action"`
	ID     string `json:"id,omitempty"`
}

func apiAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var result CommandResult
	switch req.Action {
	case "start":
		if daemonRunning() {
			result = CommandResult{Output: "smtpd is already running.", Success: true}
		} else {
			result = runCmd("smtpd")
		}
	case "stop":
		result = smtpctl("stop")
	case "restart":
		smtpctl("stop")
		time.Sleep(500 * time.Millisecond)
		result = runCmd("smtpd")
	case "reload":
		result = smtpctl("reload")
	case "status":
		if !daemonRunning() {
			result = CommandResult{Output: "smtpd is NOT running (or insufficient permissions).", Success: false}
		} else {
			result = smtpctl("show", "stats")
		}
	case "show-stats":
		result = smtpctl("show", "stats")
	case "show-queue":
		result = smtpctl("show", "queue")
	case "log-brief":
		result = smtpctl("log", "brief")
	case "log-verbose":
		result = smtpctl("log", "verbose")
	case "log-trace":
		result = smtpctl("log", "trace")
	case "queue-flush":
		result = smtpctl("schedule", "all")
	case "queue-hold":
		result = smtpctl("pause", "mta")
	case "queue-resume":
		result = smtpctl("resume", "mta")
	case "bounce":
		if req.ID == "" {
			result = CommandResult{Output: "Error: no message ID provided.", Success: false}
		} else {
			result = smtpctl("bounce", req.ID)
		}
	case "remove":
		if req.ID == "" {
			result = CommandResult{Output: "Error: no message ID provided.", Success: false}
		} else {
			result = smtpctl("remove", req.ID)
		}
	case "schedule":
		if req.ID == "" {
			result = CommandResult{Output: "Error: no message ID provided.", Success: false}
		} else {
			result = smtpctl("schedule", req.ID)
		}
	case "pause-smtp":
		result = smtpctl("pause", "smtp")
	case "resume-smtp":
		result = smtpctl("resume", "smtp")
	case "pause-mta":
		result = smtpctl("pause", "mta")
	case "resume-mta":
		result = smtpctl("resume", "mta")
	case "check-config":
		result = runCmd("smtpd", "-n")
	case "show-config":
		result = runCmd("smtpd", "-nv")
	default:
		result = CommandResult{Output: fmt.Sprintf("Unknown action: %q", req.Action), Success: false}
	}
	json.NewEncoder(w).Encode(result)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running":  daemonRunning(),
		"hostname": hostname,
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	})
}

// ── Main ───────────────────────────────────────────────────────────────────────
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/logout", handleLogout)
	mux.HandleFunc("/", requireAuth(serveIndex))
	mux.HandleFunc("/api/action", requireAuth(apiAction))
	mux.HandleFunc("/api/status", requireAuth(apiStatus))
	mux.HandleFunc("/api/logs", requireAuth(handleLogStream))

	fmt.Printf("smtpd-web listening on http://localhost%s\n", listenAddr)
	fmt.Printf("log file  : %s\n", logPath)
	fmt.Printf("username  : %s\n", adminUser)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		fmt.Printf("fatal: %v\n", err)
	}
}
