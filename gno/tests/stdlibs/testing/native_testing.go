package testing

import "time"

func X_unixNano() int64 {
	return time.Now().UnixNano()
}
