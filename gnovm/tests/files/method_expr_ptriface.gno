package main

type I interface{ M() int }

func main() {
	f := (*I).M
	println(f == nil)
}

// Error:
// main/method_expr_ptriface.gno:6:7-13: (*I).M undefined (type *main.I is pointer to interface, not interface)

// TypeCheckError:
// main/method_expr_ptriface.gno:6:12: (*I).M undefined (type *I is pointer to interface, not interface)
