package main

func main() {
	var i any
	i = [1]int{1}
	println(&i.([1]int)[0])
}

// Output:
// &(1 int)

// TypeCheckError:
// main/files/addressable_6c_err.gno:6:11: invalid operation: cannot take address of i.([1]int)[0] (value of type int)
