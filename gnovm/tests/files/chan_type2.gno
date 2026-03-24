// channel type in function signature is rejected at preprocess time.
package main

func foo(ch chan int) {
}

func main() {
	foo(nil)
}

// Error:
// main/chan_type2.gno:4:13-21: channel type is not supported
