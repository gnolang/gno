// PKGPATH: gno.land/r/fvrecvorder
package fvrecvorder

type U struct{ N int }

func (u U) Plus(k int) int { return u.N + k }

func set(u *U) int { u.N = 100; return 0 }

func main(cur realm) {
	u := U{N: 1}
	println(u.Plus(set(&u)))
	a := [1]U{{N: 2}}
	println(a[0].Plus(set(&a[0])))
}

// Output:
// 100
// 100
