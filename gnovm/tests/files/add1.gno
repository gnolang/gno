package main

func main() {
	b := 2
	var a any = 5 + b
	println(a.(int))
}

// Output:
// 7
