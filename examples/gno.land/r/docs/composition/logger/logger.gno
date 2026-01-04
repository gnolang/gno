package logger

import "time"

var logs []string

// Log adds a new entry to the log
func Log(message string) {
	timestamp := time.Now().Format("15:04:05")
	logs = append(logs, "["+timestamp+"] "+message)
}

// GetLogs returns all log entries
func GetLogs() []string {
	return logs
}

// Clear removes all log entries
func Clear() {
	logs = nil
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
