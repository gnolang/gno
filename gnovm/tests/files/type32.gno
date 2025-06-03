package main

type S string

func main() {
	a := "wqe"
	b := S("qwe")
	if false {
		println(a + ":" + b)
	}
	println("done")
}

// Error:
// main/type32.gno:9:11-22: invalid operation: a<VPBlock(2,0)> + (const (":" string)) + b<VPBlock(2,1)> (mismatched types string and main.S)

// TypeCheckError:
// main/type32.gno:9:11: invalid operation: a + ":" + b (mismatched types string and S)
