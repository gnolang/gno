package main

type S struct {
	a int
}

func main() {
	var i interface{}
	i = S{a: 9}
	println(&i.(S).a)
}

// Error:
// main/files/addressable_6b_err.gno:10:10: cannot take address of i<VPBlock(1,0)>.(S<VPBlock(3,0)>).a
