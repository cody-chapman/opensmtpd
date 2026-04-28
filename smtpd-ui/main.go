package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// CommandResult holds output from running a system command
type CommandResult struct {
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Code    int    `json:"code"`
}

// ActionRequest is the JSON body for POST /api/action
type ActionRequest struct {
	Action string `json:"action"`
	ID     string `json:"id,omitempty"`
}

func runCmd(name string, args ...string) CommandResult {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = 1
		}
	}
	combined := strings.TrimSpace(stdout.String())
	if errOut := strings.TrimSpace(stderr.String()); errOut != "" {
		if combined != "" {
			combined += "\n" + errOut
		} else {
			combined = errOut
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

func smtpctl(args ...string) CommandResult {
	return runCmd("smtpctl", args...)
}

func daemonRunning() bool {
	cmd := exec.Command("smtpctl", "show", "stats")
	return cmd.Run() == nil
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
	// Daemon
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

	// Monitoring
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

	// Log level
	case "log-brief":
		result = smtpctl("log", "brief")
	case "log-verbose":
		result = smtpctl("log", "verbose")
	case "log-trace":
		result = smtpctl("log", "trace")

	// Queue
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

	// Flow
	case "pause-smtp":
		result = smtpctl("pause", "smtp")
	case "resume-smtp":
		result = smtpctl("resume", "smtp")
	case "pause-mta":
		result = smtpctl("pause", "mta")
	case "resume-mta":
		result = smtpctl("resume", "mta")

	// Config
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
	running := daemonRunning()
	status := map[string]interface{}{
		"running":  running,
		"hostname": func() string { out, _ := exec.Command("hostname").Output(); return strings.TrimSpace(string(out)) }(),
		"time":     time.Now().Format("2006-01-02 15:04:05"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/api/action", apiAction)
	mux.HandleFunc("/api/status", apiStatus)

	addr := ":8080"
	fmt.Printf("smtpd-web manager listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("fatal: %v\n", err)
	}
}
