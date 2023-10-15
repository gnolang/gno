package main

type mychan chan int

// chan int is unmamed
func main() {
	var n mychan = nil
	var u chan int = nil
	n = u

	println(n)
}

// Output:
// (nil main.mychan)
