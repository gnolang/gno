package main

type ex int

func (ex) Error() string { return "" }

type i interface {
	Error() string
}

func main() {
	r := []int(nil)
	e, ok := r.(ex)
	println(e, ok)
}

// Error:
// *invalid operation: r<VPBlock(1,0)> (variable of type []int) is not an interface*
