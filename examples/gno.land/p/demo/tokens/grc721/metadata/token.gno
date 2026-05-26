package metadata

import (
	"chain"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/bptree/v0"
)

// NewPrivateLedger attaches a metadata ledger to an already-constructed core
// token. tok must be a non-nil *grc721.Token (typically returned from
// grc721.NewToken); the ledger holds a pointer to tok solely to validate tid
// existence on writes and does NOT own the core ledger.
//
// Typical call inside a realm's crossing function:
//
//	tok, coreLedger := grc721.NewToken(0, cur, "FOO", "FOO")
//	mLedger := metadata.NewPrivateLedger(tok)
//	rLedger := royalty.NewPrivateLedger(tok, 1000) // can coexist
//
// Do not export the returned *PrivateLedger.
func NewPrivateLedger(tok *grc721.Token) *PrivateLedger {
	if tok == nil {
		panic("Token cannot be nil")
	}

	return &PrivateLedger{
		token:     tok,
		tokenURIs: bptree.NewBPTree32(),
		entries:   bptree.NewBPTree32(),
	}
}

// TokenURI returns the URI of tid (EIP-721 metadata extension).
//
// Returns grc721.ErrInvalidTokenId if tid has no URI set or does not exist.
func (led *PrivateLedger) TokenURI(tid grc721.TokenID) (string, error) {
	uri, found := led.tokenURIs.Get(tid.String())
	if !found {
		return "", grc721.ErrInvalidTokenId
	}

	return uri.(string), nil
}

// TokenMetadata returns the on-chain metadata for tid.
//
// Returns grc721.ErrInvalidTokenId if tid does not exist, or
// ErrMetadataNotSet if metadata has not been written for tid.
func (led *PrivateLedger) TokenMetadata(tid grc721.TokenID) (Metadata, error) {
	if _, err := led.token.OwnerOf(tid); err != nil {
		return Metadata{}, err
	}

	v, found := led.entries.Get(tid.String())
	if !found {
		return Metadata{}, ErrMetadataNotSet
	}

	return v.(Metadata), nil
}

// SetTokenURI stores the URI for tid.
//
// Validity invariants:
//   - tid exists on the underlying core token
func (led *PrivateLedger) SetTokenURI(tid grc721.TokenID, uri grc721.TokenURI) error {
	if _, err := led.token.OwnerOf(tid); err != nil {
		return err
	}

	led.tokenURIs.Set(tid.String(), uri.String())
	return nil
}

// Burn removes any tokenURI and Metadata associated with tid from this
// extension ledger. It is the host realm's responsibility to call Burn on
// each attached extension ledger right after grc721.PrivateLedger.Burn so
// that stale per-tid state does not linger and resurface if the same tid is
// minted again.
//
// Burn is a no-op if tid has no URI or metadata entry.
func (led *PrivateLedger) Burn(tid grc721.TokenID) {
	tidStr := tid.String()
	led.tokenURIs.Remove(tidStr)
	led.entries.Remove(tidStr)
}

// SetTokenMetadata stores metadata for tid.
//
// Validity invariants:
//   - tid exists on the underlying core token
func (led *PrivateLedger) SetTokenMetadata(tid grc721.TokenID, m Metadata) error {
	if _, err := led.token.OwnerOf(tid); err != nil {
		return err
	}

	led.entries.Set(tid.String(), m)

	chain.Emit(
		MetadataUpdateEvent,
		"token", led.token.ID(),
		"tokenId", tid.String(),
	)
	return nil
}
