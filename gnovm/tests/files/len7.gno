package main

func main() {
	println(len(new(int)))
}

// Error:
// unexpected type for len(): *int
