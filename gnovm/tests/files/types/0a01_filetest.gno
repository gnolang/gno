package main

func main() {
	println(nil == nil) // xt: <nil>, dt: <nil>
}

// Error:
// main/files/types/0a01_filetest.gno:4: invalide operation, <nil> can only be compared to hasNil
