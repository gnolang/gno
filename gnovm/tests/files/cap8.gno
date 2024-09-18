package main

func main() {
	var i *int
	println("cap", cap(i))
}

// Error:
// unexpected type for cap(): *int
