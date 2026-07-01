package main

// Slice as map[interface{}] key: must be a Gno-catchable panic naming
// the dynamic type, matching Go's runtime "hash of unhashable type".
func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	m := map[interface{}]int{}
	m[[]int{}] = 1
}

// Output:
// recover: runtime error: hash of unhashable type []int
