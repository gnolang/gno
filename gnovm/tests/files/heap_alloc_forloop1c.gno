package main

import "fmt"

var s1 []*int

func f(j int) int {
	return j + 2 // Just as an example, increment j by 2.
}

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d \n", i, *e)
		}
	}()

	for i := 0; i < 10 && i != 3; i = i + 1 {
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// This does make 'i' NameExprTypeHeapDefine.
// You can tell by the preprocess printout of i<!~...> and i<~...>.

// Preprocessed:
// file{ package main; import fmt fmt; var s1<!VPBlock(2,0)> []*((const-type int)); func f(j (const-type int))  (const-type int) { return j<VPBlock(1,0)> + (const (2 int)) }; func forLoopRef() { defer func func(){ for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range s1<VPBlock(5,0)> { fmt<VPBlock(4,0)>.Printf((const ("s1[%d] is: %d \n" string)), (const-type gonative{interface {}})(i<VPBlock(1,0)>), (const-type gonative{interface {}})(*(e<VPBlock(1,1)>))) } }(); for i<!~VPBlock(1,0)> := (const (0 int)); i<~VPBlock(1,0)> < (const (10 int)) && i<~VPBlock(1,0)> != (const (3 int)); i<;;~VPBlock(1,0)> = i<~VPBlock(1,0)> + (const (1 int)) { s1<VPBlock(4,0)> = (const (append func(x []*int,args ...*int)(res []*int)))(s1<VPBlock(4,0)>, &(i<~VPBlock(1,0)>)) } }; func main() { forLoopRef<VPBlock(3,2)>() } }

// Output:
// s1[0] is: 0
// s1[1] is: 1
// s1[2] is: 2
