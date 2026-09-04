// Package metadata is a stackable GRC721 extension for EIP-721 tokenURI and
// OpenSea on-chain metadata. Writes are issuer-only so a holder cannot rewrite metadata after acquisition.
package metadata

import (
	"errors"

	"gno.land/p/demo/tokens/grc721"
	"gno.land/p/nt/avl/v0"
)

const (
	MetadataUpdateEvent = "MetadataUpdate"
	TokenURIUpdateEvent = "TokenURIUpdate"
)

var (
	ErrTokenNotMinted = errors.New("token is not minted")
	ErrNoMetadata     = errors.New("token has no metadata")
	ErrNoTokenURI     = errors.New("token has no uri")
)

type Metadata struct {
	core *grc721.Token
	st   *storage
}

type Ledger struct {
	core *grc721.Token
	st   *storage
}

// Data is the OpenSea on-chain metadata for a single token.
type Data struct {
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

type Trait struct {
	DisplayType string
	TraitType   string
	Value       string
}

type storage struct {
	uris avl.Tree // grc721.TokenID -> string
	data avl.Tree // grc721.TokenID -> Data
}

var _ grc721.Extension = (*Ledger)(nil)

var _ grc721.ExtensionView = (*Metadata)(nil)
