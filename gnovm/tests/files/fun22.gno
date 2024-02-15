package main

import "time"

func main() {
	time.Date()
}

// Go_Error:
// files/fun22.gno:10:2: not enough arguments in call to time.Date
// 	have ()
// 	want (int, time.Month, int, int, int, int, int, *time.Location)

// Error:
// main/files/fun22.gno:6: wrong argument count in call to time<VPBlock(2,0)>.Date; want 8 got 0
