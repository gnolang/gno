package main

import "fmt"

var s1 []*int

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d\n", i, *e)
		}
	}()

	for i := 0; i < 3; i++ {
		z := i
		s1 = append(s1, &z)
		z++
	}
}

func main() {
	forLoopRef()
}

// This does make 'z' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of z<!~...> and z<~...>.

// Output:
// s1[0] is: 1
// s1[1] is: 2
// s1[2] is: 3
