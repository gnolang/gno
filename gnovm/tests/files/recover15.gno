package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	var i any = "hello"
	_ = i.(int) // Panics because i holds a string, not an int
}

// Output:
// recover: string is not of type int
