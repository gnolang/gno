package main

func main() {
	println(nil == nil) // xt: <nil>, dt: <nil>
}

// Error:
// main/files/types/0a01_filetest.gno:4: invalid operation, nil can not be compared to <nil>
