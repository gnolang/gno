package collection

import (
	"sort"

	"gno.land/p/demo/tokens/grc721"
)

// Collection is the read-only model (core token + extension read-views, no admin surface),
// safe to store in a registry or hand across realms.
type Collection struct {
	token *grc721.Token
}

// NewCollection builds a Collection over a ledger's token. Views are answered from
// the token, so a stored Collection also reports extensions registered later.
func NewCollection(ledger *grc721.PrivateLedger) *Collection {
	if ledger == nil {
		panic("collection: nil ledger")
	}

	token := ledger.ReadToken()
	if token == nil {
		panic("collection: ledger has no token")
	}

	return &Collection{token: token}
}

func (c *Collection) Token() *grc721.Token { return c.token }

// Extension returns the published read-view of the given kind, or (nil, false).
// Callers type-assert it to the concrete type they expect (e.g. *royalty.Royalty).
func (c *Collection) Extension(kind string) (grc721.ExtensionView, bool) {
	return c.token.ExtensionView(kind)
}

func (c *Collection) HasExtension(kind string) bool {
	_, ok := c.token.ExtensionView(kind)

	return ok
}

// Kinds returns the published kinds, sorted; a kind registered without a view is omitted.
func (c *Collection) Kinds() []string {
	ks := []string{}

	for _, kind := range c.token.ExtensionKinds() {
		if _, ok := c.token.ExtensionView(kind); ok {
			ks = append(ks, kind)
		}
	}

	sort.Strings(ks)

	return ks
}
