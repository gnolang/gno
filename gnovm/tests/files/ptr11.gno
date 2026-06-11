package main

type S struct{}

func (s S) String() string { return "hey" }

func main() {
	v := (*S)(nil)
	v.String()
}

// Error:
// runtime error: nil pointer dereference
