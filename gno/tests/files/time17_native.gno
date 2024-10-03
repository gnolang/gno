package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now()
	now.In(nil)
}

// Error:
// time: missing Location in call to Time.In

// Stacktrace:
// now<VPBlock(1,0)>.In(gonative{*time.Location})
//     gofunction:func(*time.Location) time.Time
// main<VPBlock(1,0)>()
//     main/files/time17_native.gno:10
