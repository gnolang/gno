// PKGPATH: gno.land/r/zrealm_wp_2
package zrealm_wp_2

var Count = 7

type Local struct{ N int }

func (l *Local) Set(cur realm, n int) { l.N = n }
func (l Local) Reset(cur realm)       { l.N = 0; Count = 0 }

var Kept = Local{N: 1}

func main(cur realm) {
	Kept.Set(cross(cur), 5)
	println("set:", Kept.N)
	Kept.Reset(cross(cur))
	println("reset:", Kept.N, Count)
}

// Output:
// set: 5
// reset: 5 0
