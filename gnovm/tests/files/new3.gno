package main

func main() {
	a := new([1]int)
	*a = [1]int{3}
	b := new([1]int)
	*b = *a
	(*a)[0] = 2
	println((*a)[0])
	println((*b)[0])
}

// Output:
// 2
// 3
