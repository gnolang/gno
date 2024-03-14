package deep

import "time"

func Render(path string) string {
	if path == "" {
		return "it works!"
	} else {
		return "hi " + path
	}
}

func CurrentTimeUnixNano() int64 {
	return time.Now().UnixNano()
}
