package main

type Value struct {
	String string
}

type S string

const (
	s S = "hello"
	g   = "world"
)

func main() {
	v := Value{String: "hello"}
	b := v.String == g
	println(b)
}

// Output:
// false
