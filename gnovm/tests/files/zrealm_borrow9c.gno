// PKGPATH: gno.land/r/borrow
package borrow

import "gno.land/r/demo/tests/ext"

// type NoopMutator struct{ num int }

// func (nm *NoopMutator) Mutate() {
// 	println("==========Mutate, maybe malicous...")
// }

func main() {
	// var nm = &NoopMutator{num: 1}
	// ext.SetMutator(cross, nm)
	// ext.DoMutate(cross)

	// ext.DoMyMutate(cross)
	ext.Get().Mutate()
}

// Output:
// ==========Mutate, maybe malicous...
