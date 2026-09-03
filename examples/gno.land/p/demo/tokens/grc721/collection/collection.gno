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

// NewCollection builds a Collection from a core token and its extension read-views.
func NewCollection(token *grc721.Token, views ...grc721.ExtensionView) *Collection {
	if token == nil {
		panic("collection: nil token")
	}

	c := &Collection{token: token}

	for _, ev := range views {
		c.attach(ev)
	}

	return c
}

func (c *Collection) attach(ev grc721.ExtensionView) {
	if ev == nil {
		return
	}

	if ev.TokenID() != c.token.ID() {
		panic("collection: view does not belong to this token")
	}

	kind := ev.ExtensionKind()
	if c.exts.Has(kind) {
		panic("collection: duplicate extension kind: " + kind)
	}

	c.exts.Set(kind, ev)
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
	return c.KindsUpTo(0)
}

// KindsUpTo returns at most limit sorted kinds. A non-positive limit returns all kinds.
func (c *Collection) KindsUpTo(limit int) []string {
	ks := []string{}

	c.exts.Iterate("", "", func(k string, _ any) bool {
		ks = append(ks, k)
		return limit > 0 && len(ks) == limit
	})

	return ks
}
