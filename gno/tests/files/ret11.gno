package main

func retVars() (a int, b int) {
	a = 1
	return
}

func main() {
	a, b := retVars()
	println(a, b)
}

// Output:
// 1 0
