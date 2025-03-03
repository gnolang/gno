package main

type s struct{}

func (s) hello() string {
	return "hello"
}

func main() {
	var v any
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
