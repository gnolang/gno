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
// file{ package main; import fmt fmt; var s1<!~VPBlock(2,0)> []*((const-type int)); func forLoopRef() { defer func func(){ for i.loopvar<!VPBlock(1,0)>, e.loopvar<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (ref(fmt) package{})).Printf((const ("s1[%d] is: %d\n" string)), i.loopvar<VPBlock(1,0)>, *(e.loopvar<VPBlock(1,1)>)) } }(); for i.loopvar<!~VPBlock(1,0)> := (const (0 int)); i.loopvar<~VPBlock(1,0)> < (const (3 int)); i.loopvar<~VPBlock(1,0)>++ { r<!VPBlock(1,1)> := i.loopvar<~VPBlock(1,0)>; r<VPBlock(1,1)>, ok<!VPBlock(1,2)> := (const (0 int)), (const (true bool)); (const (println func(...interface {})))(ok<VPBlock(1,2)>, r<VPBlock(1,1)>); s1<~VPBlock(4,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(4,0)>, &(i.loopvar<~VPBlock(1,0)>)) } }; func main() { forLoopRef<VPBlock(3,1)>() } }

// Output:
// true 0
// true 0
// true 0
// s1[0] is: 3
// s1[1] is: 3
// s1[2] is: 3
