package main

func main() {
	println(len(new(int)))
}

// Error:
// main/len7.gno:4:10-23: invalid argument: *int for built-in len

// TypeCheckError:
// main/len7.gno:4:14: invalid argument: new(int) (value of type *int) for built-in len
