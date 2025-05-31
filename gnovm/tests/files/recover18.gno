package main

func main() {
	defer func() {
		r := recover()
		println("recover:", r)
	}()

	var m map[string]int // nil map
	m["key"] = 42        // Panics when trying to assign to a nil map
}

// Output:
// recover: uninitialized map index
