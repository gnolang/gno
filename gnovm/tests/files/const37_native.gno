package main

import "time"

func main() {
	const v = time.UTC
}

// Error:
// main/files/const37_native.gno:6:8: time<VPBlock(2,0)>.UTC (variable of type gonative{*time.Location}) is not constant