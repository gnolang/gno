// channel type in variable declaration is rejected at preprocess time.
package main

func main() {
	var ch chan int
	_ = ch
}

// Error:
// main/chan_type0.gno:5:9-17: channel type is not supported
