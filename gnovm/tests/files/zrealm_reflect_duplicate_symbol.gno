// PKGPATH: gno.land/r/demo/reflectdup
package reflectdup

// The shape of gnolang/gno#6026, with a local token type so the primitive can
// be judged without a grc20 redesign attached.
//
// A realm issues two independent tokens and names them both "FOO". Everything
// the realm chooses about them is identical, so nothing the realm chooses can
// tell their events apart. The VM's ObjectID can, because the realm chooses no
// part of it.

import (
	"chain/reflect"
)

type token struct {
	symbol string
	supply int64
}

// name is the readable half: provenance, available from construction, and
// deliberately not unique. A realm may issue two tokens under one symbol.
func (t *token) name() string { return "gno.land/r/demo/reflectdup." + t.symbol }

// id is the unique half. It is empty until the token is persisted, and the
// caller has to deal with that rather than record a blank.
func (t *token) id() (reflect.ObjectID, bool) { return reflect.ObjectIDOf(t) }

// mustID is the strict form: a realm that will not record an unidentified
// token refuses instead, which pushes the work into a second call.
func (t *token) mustID() reflect.ObjectID {
	id, ok := reflect.ObjectIDOf(t)
	if !ok {
		panic("token is not persisted yet: emit its id from a later call")
	}
	return id
}

var first, second *token

func init() {
	first = &token{symbol: "FOO"}
	second = &token{symbol: "FOO"}

	// Both are still inside the stamping window here. A realm that emitted an
	// event now would have to emit it without an id, and know that it did.
	_, firstOK := first.id()
	_, secondOK := second.id()
	if firstOK || secondOK {
		panic("tokens created in init cannot be stamped yet")
	}
}

func main(cur realm) {
	// init returned, so both tokens are persisted and stamped.
	firstID := first.mustID()
	secondID := second.mustID()

	// The realm named them identically, and that is allowed.
	println("same readable name:", first.name() == second.name())
	println("name:", first.name())

	// The VM did not. This is the property #6026 needs: two tokens one realm
	// issued under one symbol are distinguishable by something the realm could
	// not have chosen, reused, or forged.
	println("distinct ids:", firstID != secondID)
	println("both ids present:", !firstID.IsZero() && !secondID.IsZero())

	// A third token, created now, is in the window the first two were in
	// during init. It has a name but no id, and mustID says so rather than
	// writing a blank into the log.
	third := &token{symbol: "FOO"}
	println("third has a name:", third.name() == first.name())
	_, thirdOK := third.id()
	println("third has an id:", thirdOK)

	println(".")
}

// Output:
// same readable name: true
// name: gno.land/r/demo/reflectdup.FOO
// distinct ids: true
// both ids present: true
// third has a name: true
// third has an id: false
// .
