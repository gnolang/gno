package main

func p() {
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()
	i := 0
	i = 1 / i
}

func main() {
	p()
}

// Output:
// runtime error: integer divide by zero
