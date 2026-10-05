// PKGPATH: gno.land/r/demo/curalias2
package curalias2

// The second rule, reached through an alias: a realm parameter after position
// 0 may not be named `cur`, however its type was spelled.

type R = realm

func g(x int, cur R) {
	_ = cur.IsCurrent()
}

func main(cur realm) {
	g(1, cur)
}

// Error:
// gno.land/r/demo/curalias2/zrealm_cur_alias_second.gno:9:1-21: only the first realm type argument of a crossing function may have name `cur`
