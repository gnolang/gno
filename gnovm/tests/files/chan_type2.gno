// channel type in function signature is rejected at parse time.
package main

func foo(ch chan int) {
}

func main() {
	foo(nil)
}

// Error:
// chan_type2.gno:4:13: channels are not permitted
