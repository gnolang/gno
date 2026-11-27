package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	s := []int{1, 2, 3}
	low := -1
	_ = s[low:2:3]
}

// Output:
// recover: runtime error: slice bounds out of range [-1::]
