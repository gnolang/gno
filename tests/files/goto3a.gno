package main

func main() {
	for i := 0; i < 1; i++ {
		goto here
	here:
		println("ok")
		return
	}
}

// Output:
// ok
