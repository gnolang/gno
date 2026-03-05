// https://github.com/gnolang/gno/issues/5233
package main

func main() {
	ch := make(chan int)
	_ = ch
}

// Error:
// main/chan_make0.gno:5:8-22: channel operations are not yet supported
