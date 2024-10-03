package main

func main() {
	var a int
	p := &a
	*p = 1
	println(*p)
}

// Output:
// 1
