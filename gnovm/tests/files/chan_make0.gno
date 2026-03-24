// https://github.com/gnolang/gno/issues/5233
// make(chan T) is rejected because channel type is not supported.
package main

func main() {
	ch := make(chan int)
	_ = ch
}

// Error:
// main/chan_make0.gno:6:13-21: channel type is not supported
