package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	arr := []int{1, 2, 3}
	_ = arr[3] // Panics because index 3 is out of bounds
}

// Output:
// recover: slice index out of bounds: 3 (len=3)
