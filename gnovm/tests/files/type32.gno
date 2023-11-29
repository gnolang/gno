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
// main/files/type32.gno:9#1: cannot use string as main.S without explicit conversion
