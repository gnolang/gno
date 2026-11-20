// PKGPATH: gno.land/r/zrealm_fv_cost
package zrealm_fv_cost

type L0 struct{ A, B, C, D int }
type L1 struct{ X, Y, Z, W L0 }
type L2 struct{ X, Y, Z, W L1 }
type L3 struct{ X, Y, Z, W L2 }

func main(cur realm) {
	var a, b L3
	a.X.X.X.A = 1
	for i := 0; i < 30000; i++ {
		b = a
		a = b
	}
	println(b.X.X.X.A)
}

// Output:
// 1
