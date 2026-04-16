package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	arr := []int{1, 2, 3}
	idx := -1
	_ = arr[idx:] // Panics because of negative index
}

// Output:
// recover: runtime error: invalid slice index -1 (index must be non-negative)
