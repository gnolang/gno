package main

import "time"

func main() {
	time.Date()
}

// Error:
// main/fun22.gno:6:2-13: wrong argument count in call to time<VPBlock(2,0)>.Date; want 8 got 0

// TypeCheckError:
// main/fun22.gno:6:12: not enough arguments in call to time.Date
// 	have ()
// 	want (int, time.Month, int, int, int, int, int, *time.Location)
