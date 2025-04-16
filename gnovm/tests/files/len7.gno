package main

func main() {
	println(len(new(int)))
}

// Error:
// main/files/len7.gno:4:10: unexpected type for len(): int
