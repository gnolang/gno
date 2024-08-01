package main

type A interface {
	Do(s string)
}

func main() {
	var a A
	_ = a.(A)
}

// Error:
// panic: interface conversion: interface is nil, not main.A
