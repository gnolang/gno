package main

import "fmt"

var p int

func main() {
	b, x := makeClosure(1)
}

func makeClosure(a int) (b int, c func()) {
	return 2, func() {
		fmt.Println(a, b, c, p)
	}
}

// Preprocessed:
// XXX this needs to capture p too.
// file{ package main; import fmt fmt; var p<!VPBlock(2,0)> (const-type int); func main() { b<!VPBlock(1,0)>, x<!VPBlock(1,1)> := makeClosure<VPBlock(3,2)>((const (1 int))) }; func makeClosure(a~ (const-type int)) b~ (const-type int), c~ func() { return (const (2 int)), func func(){ fmt<VPBlock(3,0)>.Println(a<~VPBlock(1,0)>, b<~VPBlock(1,1)>, c<~VPBlock(1,2)>, p<VPBlock(4,0)>) }<a<()~VPBlock(1,0)>, b<()~VPBlock(1,1)>, c<()~VPBlock(1,2)>> } }
