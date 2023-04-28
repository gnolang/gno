package main

func main() {
	for {
		goto here
	here:
		println("ok")
		return
	}
}

// Output:
// ok
