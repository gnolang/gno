package main

func main() {
	var x any = [2]int{1, 1}
	s := x.([2]int)[:]
	s[0] = 9
	println(x.([2]int)[0])
}

// Error:
// main/addressable_12j_err.gno:5:7-20: cannot slice unaddressable value x<VPBlock(1,0)>.([(const (2 int))](const-type int))

// TypeCheckError:
// main/addressable_12j_err.gno:5:7: cannot slice unaddressable value x.([2]int) (comma, ok expression of type [2]int)
