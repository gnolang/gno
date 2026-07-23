// Package logger is a sub-component of /r/docs/composition.
// It holds an append-only log that only the composition realm can change.
package logger

import (
	"time"
)

var logs []string

// allowedCaller is the only realm allowed to modify the logger
const allowedCaller = "gno.land/r/docs/composition"

// Log adds a new entry to the log (restricted to composition realm)
func Log(cur realm, message string) {
	assertAuthorized(cur)
	timestamp := time.Now().Format("15:04:05")
	logs = append(logs, "["+timestamp+"] "+message)
}

// Clear removes all log entries (restricted to composition realm)
func Clear(cur realm) {
	assertAuthorized(cur)
	logs = nil
}

// assertAuthorized ensures only the composition realm can modify state
func assertAuthorized(cur realm) {
	caller := cur.Previous()
	if caller.PkgPath() != allowedCaller {
		panic("unauthorized: only " + allowedCaller + " can modify logger")
	}
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
