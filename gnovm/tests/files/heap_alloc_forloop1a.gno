package main

import "fmt"

type Int int

var s1 []*Int

func inc2(j *Int) {
	*j = *j + 2 // Just as an example, increment j by 2.
}

func forLoopRef() {
	defer func() {
		for i, e := range s1 {
			fmt.Printf("s1[%d] is: %d\n", i, *e)
		}
	}()

	for i := Int(0); i < 10; inc2(&i) {
		s1 = append(s1, &i)
	}
}

func main() {
	forLoopRef()
}

// Preprocessed:
// file{ package main; import fmt fmt; type Int (const-type main.Int); var s1<!~VPBlock(2,1)> []*(typeval{main.Int}); func inc2(j *(typeval{main.Int})) { *(j<VPBlock(1,0)>) = *(j<VPBlock(1,0)>) + (const (2 main.Int)) }; func forLoopRef() { defer func func(){ for i.loopvar<!VPBlock(1,0)>, e.loopvar<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (ref(fmt) package{})).Printf((const ("s1[%d] is: %d\n" string)), i.loopvar<VPBlock(1,0)>, *(e.loopvar<VPBlock(1,1)>)) } }(); for i.loopvar<!~VPBlock(1,0)> := (const (0 main.Int)); i.loopvar<~VPBlock(1,0)> < (const (10 main.Int)); inc2<VPBlock(4,2)>(&(i.loopvar<~VPBlock(1,0)>)) { s1<~VPBlock(4,1)> = (const (append func([]*main.Int, ...*main.Int) []*main.Int))(s1<~VPBlock(4,1)>, &(i.loopvar<~VPBlock(1,0)>)) } }; func main() { forLoopRef<VPBlock(3,3)>() } }

// Output:
// s1[0] is: 0
// s1[1] is: 2
// s1[2] is: 4
// s1[3] is: 6
// s1[4] is: 8
