package logger

import (
	"chain/runtime"
	"time"
)

var logs []string

// allowedCaller is the only realm allowed to modify the logger
const allowedCaller = "gno.land/r/docs/composition"

// Log adds a new entry to the log (restricted to composition realm)
func Log(_ realm, message string) {
	assertAuthorized()
	timestamp := time.Now().Format("15:04:05")
	logs = append(logs, "["+timestamp+"] "+message)
}

// Clear removes all log entries (restricted to composition realm)
func Clear(_ realm) {
	assertAuthorized()
	logs = nil
}

// assertAuthorized ensures only the composition realm can modify state
func assertAuthorized() {
	caller := runtime.PreviousRealm()
	if caller.PkgPath() != allowedCaller {
		panic("unauthorized: only " + allowedCaller + " can modify logger")
	}
}

// GetLogs returns all log entries
func GetLogs() []string {
	return logs
}

// RenderLog returns a markdown snippet showing the log entries
func RenderLog() string {
	if len(logs) == 0 {
		return "_No activity logged yet._"
	}

	out := ""
	// Show last 10 entries, most recent first
	start := len(logs) - 10
	if start < 0 {
		start = 0
	}
	for i := len(logs) - 1; i >= start; i-- {
		out += "- " + logs[i] + "\n"
	}
	return out
}

func Render(_ string) string {
	return "# Logger Component\n\n" + RenderLog() + "\n\nThis is a sub-component of [/r/docs/composition](/r/docs/composition)."
}
