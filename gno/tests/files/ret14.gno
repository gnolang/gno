package main

func retVars() (a int, b int) {
	for {
		defer func() {
			// a = 2
			b = 2
		}()
		return 2, 1
	}
}

func main() {
	a, b := retVars()
	println(a, b)
}

// Output:
// 2 2
