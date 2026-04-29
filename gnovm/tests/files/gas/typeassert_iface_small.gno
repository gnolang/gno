// MAXALLOC: 500000000
package main

// Type assertion to interface charges gas per interface method.
//
// typeassert_iface_large.gno: same with 10 methods.
type I interface {
	M0()
}

type S struct{}

func (S) M0() {}

func main() {
	var x interface{} = S{}
	_ = x.(I)
}

// Gas:
// 5007
