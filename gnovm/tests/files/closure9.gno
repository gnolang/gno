package main

import "fmt"

var p int

func main() {
	b, x := makeClosure(1)
	x(true)
}

func makeClosure(a int) (b int, c func(bool)) {
	return 2, func(recurse bool) {
		fmt.Println(a, b, p)
		if recurse {
			c(false)
		}
	}
}

// Preprocessed:
// file{ package main; import fmt fmt; var p<!~VPBlock(2,0)> (const-type int); func main() { b<!VPBlock(1,0)>, x<!VPBlock(1,1)> := makeClosure<VPBlock(3,2)>((const (1 int))); x<VPBlock(1,1)>((const (true bool))) }; func makeClosure(a~ (const-type int)) b~ (const-type int), c~ func(.arg_0 (const-type bool)) { return (const (2 int)), func func(recurse (const-type bool)){ (const (ref(fmt) package{})).Println(a<~VPBlock(1,1)>, b<~VPBlock(1,2)>, (const (ref(main) package{})).p); if recurse<VPBlock(2,0)> { c<~VPBlock(2,3)>((const (false bool))) } }<a<()~VPBlock(1,0)>, b<()~VPBlock(1,1)>, c<()~VPBlock(1,2)>> } }

// Output:
// 1 2 0
// 1 2 0

// TypeCheckError:
// main/files/closure9.gno:8:2: declared and not used: b
