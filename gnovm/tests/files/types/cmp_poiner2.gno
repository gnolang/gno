package main

func main() {
	a := []int{1, 2}
	b := []int{1, 2}
	println(&a == &b)
}

// Output:
// false
