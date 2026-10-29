package main

func main() {
	i := new(int)
	println("cap", cap(i))
}

// Error:
// main/cap9.gno:5:17-23: invalid argument: *int for built-in cap

// TypeCheckError:
// main/cap9.gno:5:21: invalid argument: i (variable of type *int) for built-in cap
