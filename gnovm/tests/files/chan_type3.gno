// channel type in struct field is rejected at preprocess time.
package main

type S struct {
	ch chan int
}

func main() {
}

// Error:
// main/chan_type3.gno:5:5-13: channel type is not supported
