package fuzzing

import "time"

// X_unixNano returns the current time as a Unix timestamp in nanoseconds.

func X_unixNano() int64 {
	return time.Now().UnixNano()
}
