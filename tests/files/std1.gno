package main

import "std"

func main() {
	timestamp := std.GetTimestamp()
	result := std.FormatTimestamp(timestamp, "Jan 2, 2006 at 3:04pm (MST)")
	println(result)
}

// Output:
// Feb 13, 2009 at 11:31pm (UTC)
