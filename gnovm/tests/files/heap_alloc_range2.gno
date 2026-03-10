package main

import "fmt"

var s1 []*int

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d\n", i, *e)
		}
	}()

	s := []int{0, 1, 2}
	for _, v := range s {
		s1 = append(s1, &v)
	}
}

func main() {
	forLoopRef()
}

// Preprocessed:
// file{ package main; import fmt fmt; var s1<!~VPBlock(2,0)> []*((const-type int)); func forLoopRef() { defer func func(){ for i.loopvar<!VPBlock(1,0)>, e.loopvar<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (ref(fmt) package{})).Printf((const ("s1[%d] is: %d\n" string)), i.loopvar<VPBlock(1,0)>, *(e.loopvar<VPBlock(1,1)>)) } }(); s<!VPBlock(1,0)> := (const-type []int){(const (0 int)), (const (1 int)), (const (2 int))}; for _<VPBlock(0,0)>, v<!~VPBlock(1,0)> := range s<VPBlock(2,0)> { s1<~VPBlock(4,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(4,0)>, &(v<~VPBlock(1,0)>)) } }; func main() { forLoopRef<VPBlock(3,1)>() } }

// Output:
// s1[0] is: 0
// s1[1] is: 1
// s1[2] is: 2
