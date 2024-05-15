package main

type S struct {
	x int
}

func main() {
	m := make(map[S]string)
	m[S{x: 0}] = "test"
	println(m)
}

// Error:
// FieldType is a pseudotype with no elements
