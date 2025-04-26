package main

func main() {
	var i *int
	println("cap", cap(i))
}

// Error:
// main/files/cap8.gno:5:17: unexpected type for cap(): *int

// TypeCheckError:
// main/files/cap8.gno:5:21: invalid argument: i (variable of type *int) for built-in cap

