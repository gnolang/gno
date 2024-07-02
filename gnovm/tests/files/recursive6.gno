package main

type SelfReferencing interface {
	Self() SelfReferencing
}

type Implementation struct {
	// Some implementation details...
}

func (impl Implementation) Self() SelfReferencing {
	return &impl
}

func main() {
	var obj Implementation
	var intf SelfReferencing = obj
	_ = intf.Self()
	println("ok")
}

// Output:
// ok
