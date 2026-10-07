package main

type S struct {
	a int
}

func main() {
	var i any
	i = S{a: 9}
	println(&i.(S).a)
}

// TypeCheckError:
// main/addressable_6b_err.gno:10:11: invalid operation: cannot take address of i.(S).a (value of type int)

// Error:
// main/addressable_6b_err.gno:10:10-18: invalid operation: cannot take address of i<VPBlock(1,0)>.(typeval{main.S}).a (value of type int)
