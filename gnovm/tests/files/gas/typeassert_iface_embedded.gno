// MAXALLOC: 500000000
package main

// Embedded interfaces are recursively counted via TotalMethodCount,
// so gas matches a flat interface with the same total method count.
//
// typeassert_iface_large.gno: equivalent flat interface with 10 methods.
type IA interface {
	M0()
	M1()
	M2()
	M3()
	M4()
}

type IB interface {
	M5()
	M6()
	M7()
	M8()
	M9()
}

type I interface {
	IA
	IB
}

type S struct{}

func (S) M0() {}
func (S) M1() {}
func (S) M2() {}
func (S) M3() {}
func (S) M4() {}
func (S) M5() {}
func (S) M6() {}
func (S) M7() {}
func (S) M8() {}
func (S) M9() {}

func main() {
	var x interface{} = S{}
	_ = x.(I)
}

// Gas:
// 6322
