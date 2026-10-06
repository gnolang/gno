package main

import "fmt"

// Interface holding a pointer value: &i should still be *interface{}.

type S struct{ X int }

func main() {
	var i interface{} = &S{X: 1}
	p := &i
	fmt.Printf("%T\n", p)
}

// Output:
// *interface {}
