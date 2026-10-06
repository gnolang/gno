// named channel type declaration is rejected at parse time.
package main

type C chan int

func main() {
}

// Error:
// chan_type1.gno:4:8: channels are not permitted
