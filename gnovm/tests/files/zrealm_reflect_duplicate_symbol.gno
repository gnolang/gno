// PKGPATH: gno.land/r/demo/reflectdup
package reflectdup

// The shape of gnolang/gno#6026, with a local token type so the primitive can
// be judged without a grc20 redesign attached.
//
// A realm issues two independent tokens and names them both "FOO". Everything
// the realm chooses about them is identical, so nothing the realm chooses can
// tell their events apart. The VM's address can, because the realm chooses no
// part of it, and because the address is an ordinary address, a holder ledger
// that has never heard of objects can already tell the two tokens apart.

import (
	"chain/reflect"
)

type Token struct {
	symbol string
	supply int64
}

// Name is the readable half: provenance, available from construction, and
// deliberately not unique. A realm may issue two tokens under one symbol.
func (t *Token) Name() string { return "gno.land/r/demo/reflectdup." + t.symbol }

// Address is the unique half, and what makes a Token satisfy
// reflect.Addressable. It exists from the first call after the token is
// persisted; before that, refusing beats handing out a blank.
func (t *Token) Address() address {
	obj, ok := reflect.Of(t)
	if !ok {
		panic("token not persisted yet: emit its address from a later call")
	}
	return obj.Address()
}

var first, second *Token

func init() {
	first = &Token{symbol: "FOO"}
	second = &Token{symbol: "FOO"}

	// Both are inside the stamping window here. A realm emitting an event now
	// would have to emit it without an address, and know that it did.
	if _, ok := reflect.Of(first); ok {
		panic("tokens created in init cannot be stamped yet")
	}
}

// supplyOf stands in for any downstream consumer: an indexer, a registry, a
// balance table. It keys by address and knows nothing about objects.
var supplyOf = map[address]int64{}

func main(cur realm) {
	// The realm named them identically, and that is allowed.
	println("first name: ", first.Name())
	println("second name:", second.Name())
	println("names collide:", first.Name() == second.Name())
	println("")

	// The VM did not. This is what #6026 needs: two tokens one realm issued
	// under one symbol, told apart by something the realm could not have
	// chosen, reused or forged.
	println("first addr: ", first.Address())
	println("second addr:", second.Address())
	println("addresses collide:", first.Address() == second.Address())
	println("")

	// So a consumer keyed by address keeps them apart with no special casing.
	supplyOf[first.Address()] = 1000
	supplyOf[second.Address()] = 42
	println("first supply: ", supplyOf[first.Address()])
	println("second supply:", supplyOf[second.Address()])
	println("distinct entries:", len(supplyOf))
	println("")

	// A third token created now is in the window the first two were in during
	// init: it has a name but no address, and Address says so rather than
	// writing a blank into the log.
	third := &Token{symbol: "FOO"}
	println("third name matches:", third.Name() == first.Name())
	_, ok := reflect.Of(third)
	println("third has an address:", ok)

	println(".")
}

// Output:
// first name:  gno.land/r/demo/reflectdup.FOO
// second name: gno.land/r/demo/reflectdup.FOO
// names collide: true
//
// first addr:  g1vfk0e46hfcgq82v30du00a4elaf7c4gvmgs43e
// second addr: g1ch9j854thst84rj8kp7nuqxw5aexju5c5hu8tv
// addresses collide: false
//
// first supply:  1000
// second supply: 42
// distinct entries: 2
//
// third name matches: true
// third has an address: false
// .
