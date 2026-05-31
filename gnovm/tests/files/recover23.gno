package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	s := []int{1, 2, 3}
	high := 5
	_ = s[0:high:3]
}

// Output:
// recover: runtime error: invalid slice index 5 > 3
