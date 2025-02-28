package main

type S struct{}

func (s S) String() string { return "ossas" }

type Stringer interface {
	String() string
}

func main() {
	v := (*S)(nil)
	Stringer(v).String()
}
