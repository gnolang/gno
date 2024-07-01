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

// Output:
