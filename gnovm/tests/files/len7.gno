package main

func main() {
	println(len(new(int)))
}

// Error:
// unexpected type for len(): *int

// TypeCheckError:
// main/files/len7.gno:4:14: invalid argument: new(int) (value of type *int) for built-in len
