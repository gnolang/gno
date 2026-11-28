// PKGPATH: gno.land/r/demo/curalias2
package curalias2

// The second rule, reached through an alias: a realm parameter after position
// 0 may not be named `cur`, however its type was spelled. Preprocess reaches
// it through a function type; the deploy gate refuses the name at every site.

type R = realm

var g func(x int, cur R)

func main(cur realm) {
	_ = g
}

// Error:
// gno.land/r/demo/curalias2/zrealm_cur_alias_second.gno:10:7-25: only the first realm type argument of a crossing function may have name `cur`

// TypeCheckError:
// gno.land/r/demo/curalias2/zrealm_cur_alias_second.gno:10:19: `cur` is a contextual keyword: it may only be declared as the first parameter of a crossing function
