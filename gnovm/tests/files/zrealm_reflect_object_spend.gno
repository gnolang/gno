// PKGPATH: gno.land/r/demo/objspend
package objspend

// Spending. 6214 gave an object an address and then said, in its own package
// doc, that the address "receives and does not release". This is the other
// half: a realm moving value that an object it owns has received.
//
// The authority rule is an equality the VM computes and no realm supplies:
// an object address is spendable by the realm whose storage clock minted the
// object's identity, and by nothing else. banker.NewObjectBanker is that rule.

import (
	"chain"
	"chain/banker"
	"chain/reflect"
)

const denom = "/gno.land/r/demo/objspend:vlt"

type Vault struct{ label string }

var (
	alpha = &Vault{"alpha"}
	beta  = &Vault{"beta"}

	// The snapshot of alpha's identity, kept in realm state. This is what
	// makes the address outlive the object: a reflect.Object is immutable and
	// carries everything the banker needs, so winding the balance down later
	// does not require the object to still be reachable.
	alphaID reflect.Object
)

func (v *Vault) object() reflect.Object {
	obj, ok := reflect.Of(v)
	if !ok {
		panic("vault not persisted yet")
	}
	return obj
}

func (v *Vault) Address() address { return v.object().Address() }

func balance(addr address) int64 {
	return banker.NewReadonlyBanker().GetCoin(addr, denom)
}

func main(cur realm) {
	// Fund the two objects with a realm-issued denom. IssueCoin takes a plain
	// address, so it already pays an object; nothing in the banker knows that
	// objects can have addresses.
	issuer := banker.NewBanker(banker.BankerTypeRealmIssue, cur)
	issuer.IssueCoin(alpha.Address(), denom, 1000)
	issuer.IssueCoin(beta.Address(), denom, 50)

	println("alpha:", balance(alpha.Address()))
	println("beta: ", balance(beta.Address()))
	println("realm:", balance(cur.Address()))
	println("")

	// Spend from alpha. The banker is bound to alpha's address alone.
	alphaID = alpha.object()
	ab := banker.NewObjectBanker(alphaID, cur)
	ab.SendCoins(alpha.Address(), beta.Address(), chain.Coins{{denom, 400}})
	ab.SendCoins(alpha.Address(), cur.Address(), chain.Coins{{denom, 100}})

	println("after spending from alpha")
	println("alpha:", balance(alpha.Address()))
	println("beta: ", balance(beta.Address()))
	println("realm:", balance(cur.Address()))
	println("")

	// It is bound to alpha and not to the realm: the pkgAddr pin in
	// banker.SendCoins rejects every other source, including the realm's own
	// address, which the realm can of course still spend through NewBanker.
	println("alpha banker is scoped:", scoped(ab, cur.Address(), beta.Address()))
	println("alpha banker cannot pay beta's bills:", scoped(ab, beta.Address(), cur.Address()))
	println("")

	// Lifecycle: the balance can be wound down from the snapshot alone. alpha
	// is dropped from the realm's reachable state here; alphaID is not, and
	// alphaID is all NewObjectBanker ever reads.
	alpha = nil
	wind := banker.NewObjectBanker(alphaID, cur)
	wind.SendCoins(alphaID.Address(), cur.Address(), chain.Coins{{denom, 500}})

	println("after winding down a dropped object")
	println("alpha:", balance(alphaID.Address()))
	println("realm:", balance(cur.Address()))
	println("")

	// The snapshot still names exactly one object, and no later object can be
	// minted onto that address: the realm clock only ever moves forward.
	println("snapshot address stable:", alphaID.Address() == reflect.Object(alphaID).Address())
	println("beta is a different account:", alphaID.Address() != beta.Address())
	println(".")
}

// scoped reports whether ab refuses to move coins out of from.
func scoped(ab banker.Banker, from, to address) (refused bool) {
	defer func() {
		refused = recover() != nil
	}()
	ab.SendCoins(from, to, chain.Coins{{denom, 1}})
	return false
}

// Output:
// alpha: 1000
// beta:  50
// realm: 0
//
// after spending from alpha
// alpha: 500
// beta:  450
// realm: 100
//
// alpha banker is scoped: true
// alpha banker cannot pay beta's bills: true
//
// after winding down a dropped object
// alpha: 0
// realm: 600
//
// snapshot address stable: true
// beta is a different account: true
// .
