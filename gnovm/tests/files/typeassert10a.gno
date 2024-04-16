package main

type A interface {
	Do(s string)
}

func main() {
	var a A
	_ = a.(A)
}

// Error:
// interface conversion: interface is nil, not main.A
