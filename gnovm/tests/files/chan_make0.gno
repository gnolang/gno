// https://github.com/gnolang/gno/issues/5233
// make(chan T) is rejected because channel type is not yet supported.
package main

func main() {
	ch := make(chan int)
	_ = ch
}

// Error:
// channel type is not yet supported
