package main

import (
	"time"
)

func main() {
	time.Now = func() time.Time {
		return time.Time{}
	}
}

// Error:
// main/files/assign29_native.gno:8:2: cannot assign to time<VPBlock(2,0)>.Now (neither addressable nor a map index expression)
