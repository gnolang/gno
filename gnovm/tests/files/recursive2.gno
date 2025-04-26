package main

type A struct {
	X B
}

type B struct {
	X C
}

type C struct {
	X A
}

func main() {
	var p, q A
	println(p == q)
}

// Error:
// main/files/recursive2.gno:3:6: invalid recursive type: A -> B -> C -> A

// TypeCheckError:
// main/files/recursive2.gno:3:6: invalid recursive type A; main/files/recursive2.gno:3:6: 	A refers to B; main/files/recursive2.gno:7:6: 	B refers to C; main/files/recursive2.gno:11:6: 	C refers to A
