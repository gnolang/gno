package adder

import (
	"strconv"
	"time"

	"gno.land/p/moul/txlink"
)

// Global variables to store the current number and last update timestamp
var (
	number     int
	lastUpdate time.Time
)

// Add function to update the number and timestamp
func Add(_ realm, n int) {
	number += n
	lastUpdate = time.Now()
}

// Render displays the current number value, last update timestamp, and a link to call Add with 42
func Render(_ string) string {
	// Display the current number and formatted last update time
	result := "# Add Example\n\n"
	result += "Current Number: " + strconv.Itoa(number) + "\n\n"
	result += "Last Updated: " + formatTimestamp(lastUpdate) + "\n\n"

	// Generate a transaction link to call Add with 42 as the default parameter
	txLink := txlink.Call("Add", "n", "42")
	result += "[Increase Number](" + txLink + ")\n"

	return result
}

// Helper function to format the timestamp for readability
func formatTimestamp(timestamp time.Time) string {
	if timestamp.IsZero() {
		return "Never"
	}
	return timestamp.Format("2006-01-02 15:04:05")
}
