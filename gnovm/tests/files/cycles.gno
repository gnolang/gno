package main

func main() {
	recurse()
}

func recurse() int {
	return recurse()
}

// Error:
// CPU cycle overrun
