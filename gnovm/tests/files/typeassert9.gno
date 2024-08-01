package main

// First interface
type Reader interface {
	Read() string
}

// Second interface
type Writer interface {
	Write() string
}

func main() {
	var reader Reader

	_ = reader.(Writer)
}

// Error:
// panic: interface conversion: interface is nil, not main.Writer
