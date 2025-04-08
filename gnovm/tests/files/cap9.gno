package main

func main() {
	i := new(int)
	println("cap", cap(i))
}

// Error:
// main/files/cap9.gno:5:17: unexpected type for cap(): *int