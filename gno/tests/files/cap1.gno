package main

func main() {
	exp := [...]string{"HELLO"}
	x := cap(&exp)
	println(x)
}

// Output:
// 1
