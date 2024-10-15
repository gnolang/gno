package main

type ex int

func (ex) Error() string { return "" }

type i interface {
	Error() string
}

func main() {
	r := []int(nil)
	e := r.(ex)
	println(e)
}

// Error:
// main/files/typeassert8.gno:13:7: invalid operation: r<VPBlock(1,0)> (variable of type []int) is not an interface
