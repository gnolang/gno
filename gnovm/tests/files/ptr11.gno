package main

type S struct{}

func (s S) String() string { return "hey" }

func main() {
	v := (*S)(nil)
	v.String()
}

// Error:
// panic: nil pointer dereference
