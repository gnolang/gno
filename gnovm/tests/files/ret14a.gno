package main

func main() {
	retVars := func() (a int, b int) {
		for {
			defer func() {
				a = 3
				b = 4
			}()
			return 1, 2 // This gets modified by defer
		}
	}

	a, b := retVars()
	println(a, b)
}

// Output:
// 3 4
