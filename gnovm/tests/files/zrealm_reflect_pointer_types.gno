// PKGPATH: gno.land/r/demo/reflectptrtypes
package reflectptrtypes

// One object reached through several pointer types is one Object. The type the
// caller converted to is a property of the route, not of the object, so it
// must not split one object into several map keys.

import (
	"chain/reflect"
	"strings"
)

type Token struct{ n int }
type Alias Token
type Mine strings.Builder

var (
	tok   = &Token{n: 1}
	b     = &strings.Builder{}
	index = map[string]int{"a": 1}
	alias = index
)

func main(cur realm) {
	a, _ := reflect.Of(tok)
	viaAlias, _ := reflect.Of((*Alias)(tok))
	viaAnon, _ := reflect.Of((*struct{ n int })(tok))
	println("same object via *Alias:   ", a == viaAlias)
	println("same object via *struct{}:", a == viaAnon)
	seen := map[reflect.Object]bool{a: true, viaAlias: true, viaAnon: true}
	println("map entries for one object:", len(seen))

	x1, _ := reflect.Of(b)
	x2, _ := reflect.Of((*Mine)(b))
	println("same object via *Mine:    ", x1 == x2)

	// A map is a reference: both variables hold the same object.
	m1, ok1 := reflect.Of(index)
	m2, ok2 := reflect.Of(alias)
	println("map has an address:       ", ok1 && ok2)
	println("both holders agree:       ", m1 == m2)

	// A stdlib's own objects are not a realm's, and get no identity.
	_, ok := reflect.Of(strings.ToUpper)
	println("stdlib func has an address:", ok, reflect.HasIdentity(strings.ToUpper))
	println(".")
}

// Output:
// same object via *Alias:    true
// same object via *struct{}: true
// map entries for one object: 1
// same object via *Mine:     true
// map has an address:        true
// both holders agree:        true
// stdlib func has an address: false false
// .
