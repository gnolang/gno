package main

type T struct{ n int }

func (t T) M() {}

func main() {
	t := T{}
	_ = &t.M
}

// TypeCheckError:
// main/addressable_12b_err.gno:9:7: invalid operation: cannot take address of t.M (value of type func())

// Error:
// main/addressable_12b_err.gno:9:6-10: invalid operation: cannot take address of t<VPBlock(1,0)>.M (value of type func())
