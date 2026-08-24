// PKGPATH: gno.land/r/decltypeembed
package decltypeembed

// Canary via method promotion. Companion to decltype.gno.
//
// Value is declared on Base, not on Outer; Outer.Methods never holds
// Value. At the call site O.Value(), the VM walks Outer's embedded-
// field chain, reaches Base, and resolves Value from Base.Methods at
// runtime. So this test is effectively a canary for Base.Methods:
// a mutation that removed Value from Base's slice or swapped its body
// would surface as a panic or wrong Output here — through a dispatch
// path (promotion) that is distinct from the direct concrete dispatch
// covered by the base test.
//
// The `// Types:` block below additionally pins that both Base and
// Outer are persisted as RefType (`{"ID": "..."}`), exercising the
// PR's encoding on an embedded-type shape.

type Base struct {
	V int
}

func (b Base) Value() int { return b.V }

type Outer struct {
	Base
}

var O = Outer{Base: Base{V: 42}}

func main(cur realm) {
	// Promoted-method call routes through the embedded-field chain
	// without Outer.Methods ever holding Value.
	println(O.Value())
}

// Output:
// 42

// Types:
// Base[gno.land/r/decltypeembed.Base]=
//     {
//         "ID": "gno.land/r/decltypeembed.Base"
//     }
// Outer[gno.land/r/decltypeembed.Outer]=
//     {
//         "ID": "gno.land/r/decltypeembed.Outer"
//     }
