package main

import "fmt"

type Int int

var s1 []*Int

//func (j *Int) inc() {
//	//return *j + 2 // Just as an example, increment j by 2.
//	*j = *j + 2 // Just as an example, increment j by 2.
//}

func inc2(j *Int) {
	*j = *j + 2 // Just as an example, increment j by 2.
}

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d \n", i, *e)
		}
	}()

	for i := Int(0); i < 10; inc2(&i) {
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Output:
// s1[0] is: 10
// s1[1] is: 10
// s1[2] is: 10
// s1[3] is: 10
// s1[4] is: 10
