package main

func main() {
	var i any
	i = 5
	println(&i.(int))
}

// Error:
// main/files/addressable_6a_err.gno:6:10: cannot take address of i<VPBlock(1,0)>.((const-type int))
