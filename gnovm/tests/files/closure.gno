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
// main/files/const24.gno:6:5: constant definition loop with a
