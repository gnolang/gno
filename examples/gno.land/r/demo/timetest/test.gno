package timetest

import (
	"std"
	"time"
)

func Render(path string) string {
	time2 := time.Now()
	tz := std.GetChainID()

	return time2.String() + " | " + tz
}
