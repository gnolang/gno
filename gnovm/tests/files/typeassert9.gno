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

// Stacktrace:
// panic interface conversion: interface is nil, not main.Writer
// main<VPBlock(1,2)>()
//     main/files/typeassert9.gno:16

// Error:
// interface conversion: interface is nil, not main.Writer
