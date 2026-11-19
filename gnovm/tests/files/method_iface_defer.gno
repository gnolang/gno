// `defer i.M()` on a non-nil interface: args bind at the defer statement, but
// dispatch/receiver resolve at the deferred call (matching Go). Covers value-
// and pointer-receiver methods. Regression — this used to crash the VM.

package main

type T struct{ x int }

func (t T) Show()   { println("show", t.x) }
func (t *T) PShow() { println("pshow", t.x) }

type I interface{ Show() }
type PI interface{ PShow() }

func main() {
	var i I = T{x: 1}
	defer i.Show() // value operand, deferred

	p := &T{x: 1}
	var pi PI = p
	defer pi.PShow() // pointer operand, deferred
	p.x = 9          // observed at the deferred call → 9
}

// Output:
// pshow 9
// show 1
