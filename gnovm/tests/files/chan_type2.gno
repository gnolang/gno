// channel type in function signature is rejected at preprocess time.
package main

func foo(ch chan int) {
}

func main() {
	foo(nil)
}

// Error:
// channel type is not yet supported
