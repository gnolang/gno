package main

type s struct{}

func (s) hello() string {
	return "hello"
}

func main() {
	var v interface{}
	v = s{}
	switch v := v.(type) {
	case interface{ hello() string }:
		println(v.hello())
	default:
		panic("should not happen")
	}
}

// Output:
// hello
