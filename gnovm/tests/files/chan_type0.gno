// channel type in variable declaration is rejected at parse time.
package main

func main() {
	var ch chan int
	_ = ch
}

// Error:
// chan_type0.gno:5:9: channels are not permitted
