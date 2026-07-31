package collection

import (
	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/seqid/v0"
)

// NFT is the host-side handle: a *Collection plus the admin ledger.
//
// SECURITY: an NFT carries admin authority (Ledger(), Mint, Burn).
// Keep it in an unexported var and register its ledger-free Collection() for discovery —
// never hand an *NFT (or its Ledger()) to untrusted callers.
type NFT struct {
	col    *Collection
	ledger *grc721.PrivateLedger
}

// New constructs a core token and wraps it in an NFT.
// Must be called from a realm context (init or a crossing function) because NewToken asserts rlm.IsCurrent().
func New(name, symbol string, id seqid.ID, rlm realm) *NFT {
	token, ledger := grc721.NewToken(name, symbol, id, rlm)
	return &NFT{col: &Collection{token: token}, ledger: ledger}
}

// Wrap builds an NFT around an already-constructed token/ledger pair.
func Wrap(token *grc721.Token, ledger *grc721.PrivateLedger) *NFT {
	if token == nil || ledger == nil {
		panic("collection: nil token or ledger")
	}

	return &NFT{col: &Collection{token: token}, ledger: ledger}
}

// Collection returns the read-only model. It is the same object the NFT manages,
// so attach every extension before registering it — a later Attach mutates it too.
func (n *NFT) Collection() *Collection { return n.col }

func (n *NFT) Token() *grc721.Token { return n.col.token }

// Ledger grants full admin authority; do not expose it to untrusted callers.
func (n *NFT) Ledger() *grc721.PrivateLedger { return n.ledger }

func (n *NFT) CallerTeller() grc721.Teller { return n.col.token.CallerTeller() }

// Attach records an extension read-view for discovery;
// the view must already be wired to this NFT's ledger via its own constructor.
// Returns the NFT for chaining.
func (n *NFT) Attach(view grc721.ExtensionView) *NFT {
	n.col.attach(view)
	return n
}

func (n *NFT) Extension(kind string) (grc721.ExtensionView, bool) {
	return n.col.Extension(kind)
}

func (n *NFT) Mint(to address, tid grc721.TokenID) error { return n.ledger.Mint(to, tid) }

func (n *NFT) Burn(tid grc721.TokenID) error { return n.ledger.Burn(tid) }
