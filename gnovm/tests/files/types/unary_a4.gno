package main

func main() {
	v := 42
	p := &v
	println(v == *p)
}

// Output:
// true
