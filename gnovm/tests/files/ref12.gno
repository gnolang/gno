package main

import "fmt"

// passing &interface_var to function expecting *interface{} (non-realm).

func setIt(p *interface{}, val interface{}) {
	*p = val
}

func main() {
	var i interface{} = 100
	setIt(&i, "changed")
	fmt.Println(i)
}

// Output:
// changed
