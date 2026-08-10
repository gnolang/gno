package main

// Struct containing a slice, used as map[interface{}] key: the panic
// must name the outer struct dynamic type, not the inner slice leaf.
type S struct{ s []int }

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()
	m := map[interface{}]int{}
	m[S{}] = 1
}

// Output:
// recover: runtime error: hash of unhashable type main.S
