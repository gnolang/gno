package main

type A interface {
	Do(s string)
}

func main() {
	defer func() {
		e := recover()
		_ = e.(A)
	}()
}

// Error:
// panic: interface conversion: interface is nil, not main.A
