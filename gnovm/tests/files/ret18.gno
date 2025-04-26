package main

func deferMixPointer() (int, *int) { // uname return value
	y := new(int)
	*y = 1
	defer func() { *y++ }()
	return 1, y
}

func main() {
	a, ptr := deferMixPointer()
	println(a, *ptr)
}

// Output:
// 1 2
