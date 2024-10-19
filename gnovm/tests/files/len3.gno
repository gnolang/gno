package main

func main() {
	exp := [...]int{1, 2, 3, 4, 5}
	ptr := &exp
	println(len(ptr))
}

// Output:
// 5
