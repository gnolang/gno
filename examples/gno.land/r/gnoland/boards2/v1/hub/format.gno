package hub

import "time"

// timeToUnix converts time to Unix epoch.
func timeToUnix(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}
