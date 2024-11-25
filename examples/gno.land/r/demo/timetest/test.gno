package timetest

import (
	"time"
)

func Render(path string) string {
	location, _ := time.LoadLocation("America/New_York") // Replace with your desired timezone
	// Create a specific time in that timezone
	year, month, day := 2024, time.November, 25
	hour, minute, second := 15, 30, 0
	nyTime := time.Date(year, month, day, hour, minute, second, 0, location)

	return nyTime.String()
}
