package main

func main() {
	var s S
	assert4(s)
}

type S struct {
	s string
}

func assert4(i interface{}) {
	_ = &((i.(S)).s)
}

// Error:
// doOpRef: expr not addressable: ( string)