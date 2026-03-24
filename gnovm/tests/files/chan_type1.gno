// named channel type declaration is rejected at preprocess time.
package main

type C chan int

func main() {
}

// Error:
// main/chan_type1.gno:4:6-16: channel type is not supported
