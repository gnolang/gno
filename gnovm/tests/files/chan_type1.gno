// named channel type declaration is rejected at preprocess time.
package main

type C chan int

func main() {
}

// Error:
// channel type is not yet supported
