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
// main/assign29_native.gno:8:2-10: cannot directly mutate time.Now from main

// TypeCheckError:
// main/assign29_native.gno:8:2: use of package time not in selector; main/assign29_native.gno:8:2: cannot assign to time.Now (neither addressable nor a map index expression)
