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

// Extension returns the published read-view of the given kind, or nil.
// Callers type-assert it to the concrete type they expect (e.g. *royalty.Royalty).
func (c *Collection) Extension(kind string) grc721.ExtensionView {
	return c.token.ExtensionView(kind)
}

func (c *Collection) HasExtension(kind string) bool {
	return c.token.ExtensionView(kind) != nil
}

// Kinds returns the published kinds, sorted; a kind registered without a view is omitted.
func (c *Collection) Kinds() []string {
	return c.KindsUpTo(0)
}

// KindsUpTo returns at most limit sorted kinds. A non-positive limit returns all kinds.
func (c *Collection) KindsUpTo(limit int) []string {
	ks := []string{}

	for _, kind := range c.token.ExtensionKindsUpTo(limit) {
		if c.token.ExtensionView(kind) != nil {
			ks = append(ks, kind)
		}
	}

	sort.Strings(ks)

	return ks
}
