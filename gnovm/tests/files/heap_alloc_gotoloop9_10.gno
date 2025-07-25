package main

import "fmt"

var s1 []*int
var s2 []*int

// intersection loop
func main() {
	defer func() {
		for i, v := range s1 {
			fmt.Printf("s1[%d] is %d\n", i, *v)
		}
		for i, v := range s2 {
			fmt.Printf("s2[%d] is %d\n", i, *v)
		}
	}()

	// counter for loop
	var c1, c2 int

LOOP_1:
	x := c1
	s1 = append(s1, &x)
	println("loop_1", c1)
	c1++

LOOP_2:
	y := c2
	s2 = append(s2, &y)
	println("loop_2", c2)
	c2++

	if c1 < 3 {
		goto LOOP_1
	}

	if c2 < 6 {
		goto LOOP_2
	}
}

// Preprocessed:
// file{ package main; import fmt fmt; var s1<!~VPBlock(2,0)> []*((const-type int)); var s2<!~VPBlock(2,1)> []*((const-type int)); func main() { defer func func(){ for i<!VPBlock(1,0)>, v<!VPBlock(1,1)> := range (const (ref(main) package{})).s1 { (const (ref(fmt) package{})).Printf((const ("s1[%d] is %d\n" string)), i<VPBlock(1,0)>, *(v<VPBlock(1,1)>)) }; for i<!VPBlock(1,0)>, v<!VPBlock(1,1)> := range (const (ref(main) package{})).s2 { (const (ref(fmt) package{})).Printf((const ("s2[%d] is %d\n" string)), i<VPBlock(1,0)>, *(v<VPBlock(1,1)>)) } }(); var c1<!VPBlock(1,0)>, c2<!VPBlock(1,1)> (const-type int); x<!~VPBlock(1,2)> := c1<VPBlock(1,0)>; s1<~VPBlock(3,0)> = (const (append func([]*int, ...*int) []*int))(s1<~VPBlock(3,0)>, &(x<~VPBlock(1,2)>)); (const (println func(...interface {})))((const ("loop_1" string)), c1<VPBlock(1,0)>); c1<VPBlock(1,0)>++; y<!~VPBlock(1,3)> := c2<VPBlock(1,1)>; s2<~VPBlock(3,1)> = (const (append func([]*int, ...*int) []*int))(s2<~VPBlock(3,1)>, &(y<~VPBlock(1,3)>)); (const (println func(...interface {})))((const ("loop_2" string)), c2<VPBlock(1,1)>); c2<VPBlock(1,1)>++; if c1<VPBlock(2,0)> < (const (3 int)) { goto LOOP_1<1,0,2> }; if c2<VPBlock(2,1)> < (const (6 int)) { goto LOOP_2<1,0,6> } } }



// Output:
// loop_1 0
// loop_2 0
// loop_1 1
// loop_2 1
// loop_1 2
// loop_2 2
// loop_2 3
// loop_2 4
// loop_2 5
// s1[0] is 0
// s1[1] is 1
// s1[2] is 2
// s2[0] is 0
// s2[1] is 1
// s2[2] is 2
// s2[3] is 3
// s2[4] is 4
// s2[5] is 5
