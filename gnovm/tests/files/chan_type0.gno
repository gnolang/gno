// channel type in variable declaration is rejected at preprocess time.
package main

func main() {
	var ch chan int
	_ = ch
}

// Error:
// channel type is not yet supported
