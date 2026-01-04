package main

type Value struct {
	String string
}

type S string

const (
	s S = "hello"
	g   = "world"
	h
)

func main() {
	v := Value{String: "hello"}
	println(v.String == g)
	println(g == h)
}

// Output:
// false
// true
