package collection

import (
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

// Collection is the read-only model (core token + extension read-views, no admin surface),
// safe to store in a registry or hand across realms.
type Collection struct {
	token *grc721.Token
	exts  avl.Tree // kind -> grc721.ExtensionView
}

// NewCollection builds a Collection from a ledger's token and every read-view
// registered on that ledger.
func NewCollection(ledger *grc721.PrivateLedger) *Collection {
	if ledger == nil {
		panic("collection: nil ledger")
	}

	c := &Collection{token: ledger.ReadToken()}
	c.sync(ledger)

	return c
}

// sync copies in the read-views the ledger publishes. Views are taken from the
// ledger rather than from a caller, so there is nothing to authenticate: a view
// cannot be forged, cannot belong to another token, and cannot be a typed nil.
func (c *Collection) sync(led *grc721.PrivateLedger) {
	for _, kind := range led.ExtensionKinds() {
		if c.exts.Has(kind) {
			continue
		}

		if view, ok := led.ExtensionView(kind); ok {
			c.exts.Set(kind, view)
		}
	}
}

func (c *Collection) Token() *grc721.Token { return c.token }

// Extension returns the attached read-view of the given kind, or (nil, false).
// Callers type-assert it to the concrete type they expect (e.g. *royalty.Royalty).
func (c *Collection) Extension(kind string) (grc721.ExtensionView, bool) {
	x := c.exts.Get(kind)
	if x == nil {
		return nil, false
	}

	return x.(grc721.ExtensionView), true
}

func (c *Collection) HasExtension(kind string) bool { return c.exts.Has(kind) }

func (c *Collection) Kinds() []string {
	ks := []string{}

	c.exts.Iterate("", "", func(k string, _ any) bool {
		ks = append(ks, k)
		return false
	})

	return ks
}
