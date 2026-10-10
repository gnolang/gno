// PKGPATH: gno.land/r/demo/reflectgov
package reflectgov

// The case this package exists for: a governance proposal that owns a fund and
// carries an executable body.
//
// Give the proposal object an address and the fund stops being a bookkeeping
// entry in a map keyed by proposal number. It becomes a place on chain where
// value sits. An indexer can list its balance, a page can deep-link to it, and
// a voter can check the money is where the proposal says it is without
// trusting the realm's own accounting. Nothing is stored to make that true:
// the address is derived from the object, so no later version of this realm
// can point it somewhere else, and no second proposal can be made to share it.

import (
	"chain/reflect"
)

type Proposal struct {
	title string
	body  func() string // the executable part, an object in its own right
}

var proposals []*Proposal

func init() {
	proposals = append(proposals,
		&Proposal{
			title: "Fund the docs rewrite",
			body:  func() string { return "docs budget released" },
		},
		&Proposal{
			title: "Raise the validator set to 21",
			body:  func() string { return "valset param updated" },
		},
	)
}

// Address is where to send GNOT for this proposal, and is what makes a
// *Proposal satisfy reflect.Addressable: the proposal's fund and the
// proposal's identity are the same thing, so they cannot drift apart.
func (p *Proposal) Address() address {
	obj, ok := reflect.Of(p)
	if !ok {
		// Inside the stamping window. A proposal created by the running call
		// has no address yet, so refusing here is better than handing back a
		// blank that somebody sends coins to.
		panic("proposal not persisted yet: its fund exists from the next call on")
	}
	return obj.Address()
}

// Links renders where a human goes to look at this proposal.
//
// ok is checked here too, even though this is only ever called on a persisted
// proposal. Dropping it would render a link to the empty address, which is not
// an account, and example code gets copied.
func (p *Proposal) Links() string {
	obj, ok := reflect.Of(p)
	if !ok {
		return "  (not persisted yet, no links)"
	}
	return "  realm: /r/demo/reflectgov:proposal/" + obj.ID() + "\n" +
		"  funds: /accounts/" + string(obj.Address())
}

// balanceOf stands in for a banker or a token ledger. The point is the
// signature: it takes an address, so it already accepts an object, with no
// knowledge that objects can have addresses. That is what "objects become
// addressable" buys, and it needs no change to the thing being called.
var ledger = map[address]int64{}

func credit(to address, amount int64) { ledger[to] += amount }

// describe accepts anything the VM can name: a user, a realm, or an object.
func describe(what reflect.Addressable) string {
	return string(what.Address())
}

func main(cur realm) {
	docs, valset := proposals[0], proposals[1]

	println("proposal:", docs.title)
	println(docs.Links())
	println("")
	println("proposal:", valset.title)
	println(valset.Links())
	println("")

	// Two proposals, two funds, chosen by neither the realm nor the proposer.
	println("funds are distinct:", docs.Address() != valset.Address())

	// Funding a proposal is an ordinary transfer to an ordinary address. The
	// ledger below knows nothing about objects.
	credit(docs.Address(), 500)
	credit(valset.Address(), 120)
	credit(docs.Address(), 250)
	println("docs fund:  ", ledger[docs.Address()])
	println("valset fund:", ledger[valset.Address()])
	println("")

	// The body is an object too, so an executable proposal body is separately
	// addressable and separately linkable from the proposal that holds it.
	// That is the difference between "proposal 4" and "this exact lambda".
	body, ok := reflect.Of(docs.body)
	println("body is addressable:", ok)
	if ok {
		println("body address:       ", body.Address())
		println("body differs from proposal:", body.Address() != docs.Address())
	}
	println("")

	// Addressable takes a user, a realm or an object without knowing which.
	println("via interface:", describe(docs) == string(docs.Address()))

	println(".")
}

// Output:
// proposal: Fund the docs rewrite
//   realm: /r/demo/reflectgov:proposal/0bc75fd1e3b7df04f1efce3f4f789ae950c27735:12
//   funds: /accounts/g14wcm9ffy0mzuxflnaxmfkhtkcg9rqf986yudwg
//
// proposal: Raise the validator set to 21
//   realm: /r/demo/reflectgov:proposal/0bc75fd1e3b7df04f1efce3f4f789ae950c27735:15
//   funds: /accounts/g1mqjz9xa72efzrh9tyzzt2tla8krm2nhf92u9gl
//
// funds are distinct: true
// docs fund:   750
// valset fund: 120
//
// body is addressable: true
// body address:        g1c3849hpgjqejr62u9fvq9wueuw4akpgwcp3puy
// body differs from proposal: true
//
// via interface: true
// .
