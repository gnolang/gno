package main

func retVars() (a int, b int) {
	for {
		a = 1
		return
	}
}

func main() {
	a, b := retVars()
	println(a, b)
}

// Output:
// 1 0
