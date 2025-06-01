package main

import "time"

func main() {
	const v = time.UTC
}

// Error:
// main/const37.gno:6:8-20: time<VPBlock(2,0)>.UTC (variable of type *time.Location) is not constant

// TypeCheckError:
// main/const37.gno:6:12: time.UTC (variable of type *time.Location) is not constant
