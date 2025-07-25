package main

func main() {
	i := new(int)
	println("cap", cap(i))
}

// Error:
// unexpected type for cap(): *int

// TypeCheckError:
// main/cap9.gno:5:21: invalid argument: i (variable of type *int) for built-in cap
