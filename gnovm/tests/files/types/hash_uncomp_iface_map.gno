package main

// Map dynamic key via interface{}: previously triggered an uncatchable
// host panic. Must be a Gno-catchable Exception now.
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	m := map[interface{}]int{}
	m[map[int]int{}] = 1
}

// Output:
// recover: runtime error: hash of unhashable type map[int]int
