package main

func main() {
	var a = int(1) + 1
	println(nil << a)
}

// Error:
// main/files/types/shift_d37.gno:5:10: operator << not defined on: nil
