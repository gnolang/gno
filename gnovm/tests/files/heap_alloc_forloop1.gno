package main

import "fmt"

var s1 []*int

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d \n", i, *e)
		}
	}()

	for i := 0; i < 3; i++ {
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Output:
//s1[0] is: 0
//s1[1] is: 1
//s1[2] is: 2
