package main

func main() {
	var a = int(1) + 1
	println(nil << a)
}

// Error:
// main/files/types2/shift_d37.gno:5: operator << not defined on: <nil>
