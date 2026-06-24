// PKGPATH: gno.land/r/decltype
package decltype

// Invariant pinned by this filetest: a DeclaredType is structurally
// final by the time SetType is called on it. The cache-hit short-circuit
// in defaultStore.SetType silently relies on this — on a same-pointer
// second call, the new code skips the backend rewrite that the old code
// performed, so any mutation to dt.Methods between the two calls would
// persist old bytes while the cached pointer carries new ones.
//
// Canary axis (positional). Seeded.M() and Aliased.N() compile to
// ValuePaths with a preprocess-bound Index; at runtime the VM reads
// dt.Methods[Index] directly. This catches positional mutations — slice
// reordering, insert-before-target, remove-before-target, or body swap
// at the indexed slot — which would cause the wrong FuncValue to be
// called and diverge Output.
//
// Realm shape. `type Alias = T` is the canonical trigger for a second
// SetType call on the same T pointer during saveNewPackageValuesAndTypes
// — exactly the call the short-circuit discards. `Seeded` and `Aliased`
// are package-level values that exercise the TypeValue save path the
// PR reshaped; the `// Types:` block below pins both as RefType
// (`{"ID": "..."}`), the form the PR promotes. A future regression in
// either the sealed-Type invariant or the TypeValue-as-ref encoding
// diverges the Output or Types block.

type T struct {
	X int
}

type Alias = T

func (t T) M() int    { return t.X * 2 }
func (t T) N() string { return "n" }

var (
	Seeded  T     = T{X: 7}
	Aliased Alias = Alias{X: 9}
)

func main(cur realm) {
	println(Seeded.M())
	println(Aliased.N())
}

// Output:
// 14
// n

// Types:
// T[gno.land/r/decltype.T]=
//     {
//         "ID": "gno.land/r/decltype.T"
//     }
// Alias[gno.land/r/decltype.T]=
//     {
//         "ID": "gno.land/r/decltype.T"
//     }
