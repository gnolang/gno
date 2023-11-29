package main

import (
	"time"
)

var twentyFourHours = time.Duration(24 * time.Hour)

func main() {
	println(twentyFourHours.Hours())
}

// Output:
// 24
