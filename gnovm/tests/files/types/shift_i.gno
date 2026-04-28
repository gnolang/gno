package main

func main() {
	var a [1024]byte
	var s uint = 1
	xx := a[1.0<<s]
	println(xx)
}


// Output:
// 0