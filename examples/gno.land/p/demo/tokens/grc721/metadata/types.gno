// Package metadata provides an on-chain metadata extension for grc721.
//
// A PrivateLedger attaches per-token tokenURI and OpenSea-style Metadata to
// a pre-constructed *grc721.Token. The ledger does not own the core token,
// so a single collection may carry multiple extensions (e.g. metadata and
// royalty) at the same time.
//
// The Metadata struct mirrors OpenSea's metadata schema for tokens whose
// metadata lives on chain rather than behind a TokenURI fetch.
package metadata

import (
	"errors"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/bptree/v0"
)

// Trait is one OpenSea-style attribute entry.
type Trait struct {
	DisplayType string
	TraitType   string
	Value       string
}

// Metadata is the on-chain metadata schema for a single token.
//
// Field semantics follow https://docs.opensea.io/docs/metadata-standards.
type Metadata struct {
	Image           string
	ImageData       string
	ExternalURL     string
	Description     string
	Name            string
	Attributes      []Trait
	BackgroundColor string
	AnimationURL    string
	YoutubeURL      string
}

// IMetadata is the on-chain metadata read interface, covering both the
// EIP-721 tokenURI surface and the OpenSea-style on-chain Metadata struct.
type IMetadata interface {
	TokenURI(tid grc721.TokenID) (string, error)
	TokenMetadata(tid grc721.TokenID) (Metadata, error)
}

// Teller is the metadata extension's read-only high-level API.
//
// Per EIP-721 metadata semantics, tokenURI and on-chain Metadata are
// collection-level data controlled by the issuing realm, not by
// individual token owners. Mutating writes therefore live on
// *PrivateLedger only; the host realm must guard SetTokenURI and
// SetTokenMetadata behind its own admin policy. This Teller surfaces
// the read views as safe-to-share.
type Teller interface {
	IMetadata
}

// PrivateLedger holds per-token tokenURI and Metadata state for the
// collection. It provides administrative functions for writing URIs and
// metadata entries.
//
// The PrivateLedger is not safe to expose publicly, as it allows direct,
// unrestricted access to all administrative functions.
type PrivateLedger struct {
	// Underlying core token; used to validate tid existence on writes.
	token *grc721.Token
	// TokenID -> string (per EIP-721 tokenURI).
	tokenURIs *bptree.BPTree
	// TokenID -> Metadata (on-chain OpenSea-style).
	entries *bptree.BPTree
}

var ErrMetadataNotSet = errors.New("metadata not set for token")

const MetadataUpdateEvent = "MetadataUpdate"

// fnTeller is the canonical Teller implementation. accountFn supplies the
// acting principal for write methods; a nil accountFn marks the Teller as
// read-only.
type fnTeller struct {
	ledger    *PrivateLedger
	accountFn func(_ int, rlm realm) address
}

var (
	_ IMetadata = (*PrivateLedger)(nil)
	_ Teller    = (*fnTeller)(nil)
)

// IsCanonicalTeller reports whether t is the canonical *fnTeller produced by
// PrivateLedger.CallerTeller / RealmTeller / ReadonlyTeller /
// ImpersonateTeller. Use this at any public entry point that accepts a
// Teller from an external caller before invoking its methods.
//
// Foreign types — including embedding-based wrappers — are rejected because
// type assertions are nominal. Mirrors grc721.IsCanonicalTeller.
func IsCanonicalTeller(t Teller) bool {
	_, ok := t.(*fnTeller)
	return ok
}
