package main

type f func()

var f1 f
var f2 f

func main() {
	println(f1 == f2)
}

// Error:
// main/files/types/0f30b_filetest.gno:9: main.f can only be compared to nil
