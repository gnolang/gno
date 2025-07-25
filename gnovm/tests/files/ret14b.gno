package main

func retVars() (_ int, _ int) {
	for {
		var a, b = 1, 2
		defer func() {
			a = 3
			b = 4
		}()
		return a, b
	}
}

func main() {
	x, y := retVars()
	println(x, y)
}

// Output:
// 1 2
