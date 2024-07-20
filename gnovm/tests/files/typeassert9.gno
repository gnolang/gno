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
// _<VPBlock(0,0)> = reader<VPBlock(1,0)>.(Writer<VPBlock(3,1)>)
//     main/files/typeassert9.gno:16
// main()
//     main/files/typeassert9.gno:13

// Error:
// interface conversion: interface is nil, not main.Writer
