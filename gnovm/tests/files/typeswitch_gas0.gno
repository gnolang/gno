// MAXALLOC: 500000000
package main

// Test gas metering for a type switch with few clauses (2 cases + default).
// Compare with typeswitch_gas1.gno which has more clauses and should cost more gas.
// This verifies that doOpTypeSwitch charges gas per clause and per case.

type I1 interface {
	A()
}

type S struct{}

func (S) A() {}

func main() {
	var x interface{} = S{}

	switch x.(type) {
	case int:
		println("int")
	case I1:
		println("I1")
	default:
		println("default")
	}
}

// Output:
// I1

// Gas:
// 8763
