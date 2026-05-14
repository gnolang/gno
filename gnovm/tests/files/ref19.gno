package main

import "fmt"

// Type assertion through interface pointer.
// *p must produce a proper interface{} for the assertion to work.

func main() {
	var i interface{} = 42
	p := &i
	v := (*p).(int)
	fmt.Println(v)
}

// Output:
// 42
