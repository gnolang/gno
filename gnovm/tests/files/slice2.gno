package main

func main() {
	_ = []string{0: "foo", 2: "bar", 2: "baz"}
}

// Error:
// duplicate index 2 in array or slice literal

// TypeCheckError:
// main/slice2.gno:4:35: duplicate index 2 in array or slice literal
