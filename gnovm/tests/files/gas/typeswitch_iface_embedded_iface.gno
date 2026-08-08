// MAXALLOC: 500000000
package main

// Type switch interface case; gas counts methods promoted from an embedded
// named interface (they live in the interface's underlying type).
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

type Outer struct{ I }

func main() {
	var x interface{} = Outer{}
	switch x.(type) {
	case I:
	}
}

// Gas:
// 6987
