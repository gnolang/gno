package main

import "fmt"

// nested pointer roundtrip: *&x where x is interface{}.

func main() {
	var i interface{} = 42
	v := *&i
	fmt.Printf("%T %v\n", v, v)
}

// Output:
// int 42
