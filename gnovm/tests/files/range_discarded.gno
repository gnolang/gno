package main

func main() {
	var arr *[10]int
	s := 0
	for i, _ := range arr {
		// used to panic trying to access arr[i]
		s += i
	}
	if s != 45 {
		println("BUG")
	}
	println("OK")
}

// Output:
// OK
