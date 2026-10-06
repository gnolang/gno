package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r) // not recoverable.
	}()

	s := []int{1, 2, 3}
	_ = s[-1:] // Panics because of negative index
}

// Error:
// main/index3.gno:10:6-12: invalid argument: index must not be negative: (-1 <untyped> bigint)

// TypeCheckError:
// main/index3.gno:10:8: invalid argument: index -1 (constant of type int) must not be negative
