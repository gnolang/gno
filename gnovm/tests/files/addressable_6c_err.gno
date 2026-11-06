package main

func main() {
	var i any
	i = [1]int{1}
	println(&i.([1]int)[0])
}

// TypeCheckError:
// main/addressable_6c_err.gno:6:11: invalid operation: cannot take address of i.([1]int)[0] (value of type int)

// Error:
// main/addressable_6c_err.gno:6:10-24: invalid operation: cannot take address of i<VPBlock(1,0)>.([(const (1 int))](const-type int))[(const (0 int))] (value of type int)
