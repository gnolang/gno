// channel type in struct field is rejected at preprocess time.
package main

type S struct {
	ch chan int
}

func main() {
}

// Error:
// channel type is not yet supported
