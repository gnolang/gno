// MAXALLOC: 500000000
package main

// Type switch interface case against a struct whose methods come entirely
// from an embedded field — gas accounts for promoted methods, not just
// directly declared ones.
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

type Inner struct{}

func (Inner) M0() {}
func (Inner) M1() {}
func (Inner) M2() {}
func (Inner) M3() {}
func (Inner) M4() {}
func (Inner) M5() {}
func (Inner) M6() {}
func (Inner) M7() {}
func (Inner) M8() {}
func (Inner) M9() {}

type Outer struct{ Inner }

func main() {
	var x interface{} = Outer{}
	switch x.(type) {
	case I:
	}
}

// Gas:
// 6370
