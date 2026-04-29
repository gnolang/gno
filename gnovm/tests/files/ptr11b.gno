package main

type S struct{}

func (s S) String() string { return "hey" }

type Stringer interface {
	String() string
}

func main() {
	v := (*S)(nil)
	println(Stringer(v).String())
}

// Error:
// runtime error: nil pointer dereference
