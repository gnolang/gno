// channel type in struct field is rejected at parse time.
package main

type S struct {
	ch chan int
}

func main() {
}

// Error:
// chan_type3.gno:5:5: channels are not permitted
