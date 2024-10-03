package main

func main() {
	_ = []string{0: "foo", 2: "bar", 2: "baz"}
}

// Error:
// duplicate index 2 in array or slice literal
