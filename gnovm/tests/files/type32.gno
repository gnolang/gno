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
// main/files/type32.gno:9:11: invalid operation: mismatched types string and main.S

// TypeCheckError:
// main/files/type32.gno:9:11: invalid operation: a + ":" + b (mismatched types string and S)
