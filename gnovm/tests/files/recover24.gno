package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	s := []int{1, 2, 3}
	_ = s[0:1:10]
}

// Output:
// recover: runtime error: slice bounds out of range [0:1:10] with capacity 3
