package main

func main() {
	x := 1024
	p := 10
	if x != 1<<p {
		println("WHAT?")
	} else {
		println("nothing")
	}
}

// Output:
// nothing
