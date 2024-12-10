package main

func main() {

}

var a = func() {
	b()
}

var b = func() {
	a()
}

// Error:
// main/files/closure.gno:7:5: constant definition loop with a
