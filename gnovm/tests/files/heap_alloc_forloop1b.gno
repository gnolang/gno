package main

import "fmt"

var s1 []*int
var a interface{} = 1

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d \n", i, *e)
		}
	}()

	for i := 0; i < 3; i++ {
		r := i
		r, ok := 0, true
		println(ok, r)
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// go 1.22 loop var is not supported for now.

// Preprocessed:
// file{ package main; import fmt fmt; var s1<!VPBlock(2,0)> []*((const-type int)); var a<!VPBlock(2,1)> interface {  } = (const (1 int)); func forLoopRef() { defer func func(){ for i<!VPBlock(1,0)>, e<!VPBlock(1,1)> := range s1<VPBlock(5,0)> { fmt<VPBlock(4,0)>.Printf((const ("s1[%d] is: %d \n" string)), (const-type gonative{interface {}})(i<VPBlock(1,0)>), (const-type gonative{interface {}})(*(e<VPBlock(1,1)>))) } }(); for i<!~VPBlock(1,0)> := (const (0 int)); i<~VPBlock(1,0)> < (const (3 int)); i<~VPBlock(1,0)>++ { r<!VPBlock(1,1)> := i<~VPBlock(1,0)>; r<VPBlock(1,1)>, ok<!VPBlock(1,2)> := (const (0 int)), (const (true bool)); (const (println func(xs ...interface{})()))(ok<VPBlock(1,2)>, r<VPBlock(1,1)>); s1<VPBlock(4,0)> = (const (append func(x []*int,args ...*int)(res []*int)))(s1<VPBlock(4,0)>, &(i<~VPBlock(1,0)>)) } }; func main() { forLoopRef<VPBlock(3,2)>() } }

// Output:
// true 0
// true 0
// true 0
// s1[0] is: 3
// s1[1] is: 3
// s1[2] is: 3
