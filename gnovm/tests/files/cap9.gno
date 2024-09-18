package main

func main() {
	i := new(int)
	println("cap", cap(i))
}

// Error:
// unexpected type for cap(): *int
