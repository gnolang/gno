package main

type S struct {
	a int
}

func main() {
	var i any
	i = S{a: 9}
	println(&i.(S).a)
}

// Output:
// &(9 int)

// TypeCheckError:
// main/files/addressable_6b_err.gno:10:11: invalid operation: cannot take address of i.(S).a (value of type int)
