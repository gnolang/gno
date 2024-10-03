package main

func main() {
	x := "qwe"
	for range x {
		goto here
	here:
		println("ok")
		return
	}
}

// Output:
// ok
