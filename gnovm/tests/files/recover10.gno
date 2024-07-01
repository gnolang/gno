package main

func main() {
	defer recover()

	panic("ahhhhh")
}

// Error:
// ahhhhh
