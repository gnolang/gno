package main

type I interface{ M() }

type T struct{}

func (T) M() {}

func main() {
	var i I = T{}
	_ = &i.M
}

// TypeCheckError:
// main/addressable_12f_err.gno:11:7: invalid operation: cannot take address of i.M (value of type func())

// Error:
// main/addressable_12f_err.gno:11:6-10: invalid operation: cannot take address of i<VPBlock(1,0)>.M (value of type func())
