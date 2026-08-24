// PKGPATH: gno.land/r/test
// Mirrored (plus restart coverage) by restart_local_type.txtar's zltsh realm.
package test

type T struct{ x int }

func (t T) Get() int { return t.x }

type Getter interface{ Get() int }

var (
	G      func() any
	GG     func() any
	D1, D2 any
)

func mk1() any {
	type S struct{ T }
	return S{T{1}}
}

func mk2() any {
	type S struct{ T }
	return S{T{2}}
}

func main(cur realm) {
	// B: no-trace escape — nothing S2-typed persists at save time, not even
	// a TypeValue (the type name is preprocess-folded, not captured; only
	// the FuncValue is saved). Reload re-mints S2, same TypeID.
	type S2 struct{ T }
	G = func() any { return S2{T{7}} }
	println(G().(Getter).Get())

	// C: type declared inside a closure literal (ParentLoc = FuncLitExpr)
	GG = func() any {
		type C struct{ T }
		return C{T{5}}
	}
	println(GG().(Getter).Get())

	// D: same local type name in two different functions
	D1, D2 = mk1(), mk2()
	println(D1.(Getter).Get(), D2.(Getter).Get())
}

// Output:
// 7
// 5
// 1 2

