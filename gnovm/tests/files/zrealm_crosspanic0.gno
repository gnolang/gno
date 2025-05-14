// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

// This tests that panics that cross realm boundaries abort the transaction.

import (
	"gno.land/r/demo/tests/crossrealm_b"
)

type StructA struct {
	A any
}

type StructB struct {
	B *StructA
}

func (sb *StructB) Set(sa *StructA) {
	sb.B = sa
}

func (sb *StructB) Panic() {
	panic("success: this panic should surface")
}

func init() {
	// save StructB{} in crossrealm_b.
	sb := &StructB{}
	cross(crossrealm_b.SetObject)(sb)
}

func main() {
	sb := crossrealm_b.GetObject().(*StructB)
	//sa := &StructA{}
	//sb.Set(sa) // attach sa to crossrealm_b.

	func() {
		defer func() {
			r := recover()
			println("fail: should not be printing anything but recovered", r)
			panic("fail: the transaction should have been aborted; this defer should not even be called")
		}()
		sb.Panic()
	}()
}

// Error:
// success: this panic should surface
