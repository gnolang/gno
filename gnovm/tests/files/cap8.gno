package main

func main() {
	var i *int
	println("cap", cap(i))
}

// Error:
// main/files/cap8.gno:5:17: unexpected type for cap(): int
