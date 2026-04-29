// MAXALLOC: 500000000
package main

// Type assertion gas grows linearly with interface method count.
//
// typeassert_iface_small.gno: same test with 1 method.
type I interface {
	M0()
	M1()
	M2()
	M3()
	M4()
	M5()
	M6()
	M7()
	M8()
	M9()
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
// 5565
