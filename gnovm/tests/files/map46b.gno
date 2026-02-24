package main

func main() {
	var m map[string]int
	delete(m, "foo")
	delete(m, "bar")
	delete(m, "")

	var m2 map[int]string
	delete(m2, 0)
	delete(m2, 42)

	println("all deletes on nil maps succeeded")
}

// Output:
// all deletes on nil maps succeeded
