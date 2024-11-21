package main

func main() {
	var s uint = 33

	var u1 bool
	u1 = 1<<s != 0
	println(u1)
}

// Output:
// true
