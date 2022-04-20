package main

import "std"

func main() {
	timestamp := std.GetTimestamp()
	result := std.FormatTimestamp(timestamp, "Jan 2, 2006 at 3:04pm (MST)")
	println(result)
}

// Output:
// Jan 1, 1970 at 12:00am (UTC)
