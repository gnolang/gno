package main

func main() {
	exp := [...]string{"HELLO"}
	x := len(&exp)
	println(x)
}

// Output:
// 1
