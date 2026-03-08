// MAXALLOC: 500000000
package main

// Test gas metering for a type switch with many clauses (7 cases + default),
// including interface cases with varying method counts.
// Compare with typeswitch_gas0.gno which has fewer clauses and should cost less gas.
// This verifies that doOpTypeSwitch charges gas proportional to:
//   - the number of clauses iterated
//   - the number of cases checked
//   - the number of interface methods verified

type I1 interface {
	A()
}

type I2 interface {
	A()
	B()
}

type I3 interface {
	A()
	B()
	C()
}

type S struct{}

func (S) A() {}
func (S) B() {}
func (S) C() {}

func main() {
	var x interface{} = S{}

	switch x.(type) {
	case int:
		println("int")
	case string:
		println("string")
	case bool:
		println("bool")
	case float64:
		println("float64")
	case I1:
		println("I1")
	case I2:
		println("I2")
	case I3:
		println("I3")
	default:
		println("default")
	}
}

// Output:
// I1

// Gas:
// 9848
