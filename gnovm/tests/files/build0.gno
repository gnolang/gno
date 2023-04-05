// A test program

//go:build ((darwin && linux) || !arm) && (go1.12 || !go1.13)
// +build darwin,linux !arm
// +build go1.12 !go1.13

package main

func main() {
	println("hello world")
}

// Output:
// hello world
