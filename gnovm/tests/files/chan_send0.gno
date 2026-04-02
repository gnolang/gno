// send statement is rejected at parse time.
package main

func main() {
	var ch interface{}
	ch <- 1
}

// TypeCheckError:
// main/chan_send0.gno:6:5: invalid operation: cannot send to non-channel interface{} ch (variable of type interface{})

// Error:
// chan_send0.gno:6:2: send statements are not permitted
