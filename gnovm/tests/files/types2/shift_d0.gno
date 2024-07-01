package main

func gen() []int {
	return nil
}

func main() {
	r := gen() == nil
	println(r)
}

// Output:
// true
