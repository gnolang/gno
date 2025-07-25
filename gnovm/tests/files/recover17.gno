package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	str := "hello"
	_ = str[10] // Panics because index 10 is out of bounds
}

// Output:
// recover: index out of range [10] with length 5
