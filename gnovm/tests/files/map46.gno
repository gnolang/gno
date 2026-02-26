package main

func main() {
	var m map[string]int
	delete(m, "foo")
	println("ok")
}

// Output:
// ok
