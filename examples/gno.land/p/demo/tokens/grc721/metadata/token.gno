package metadata

import (
	"chain"

	"gno.land/p/demo/tokens/grc721"
)

const Kind = "metadata"

// NewMetadata attaches a metadata extension and registers it as a core Extension.
// Attach it before the first mint so every token is indexed.
func NewMetadata(coreLedger *grc721.PrivateLedger) (*Metadata, *Ledger) {
	if coreLedger == nil {
		panic("metadata: nil core ledger")
	}

	st := &storage{}
	core := coreLedger.ReadToken()
	led := &Ledger{core: core, st: st}
	view := &Metadata{core: core, st: st}
	coreLedger.RegisterExtension(led, view)

	return view, led
}

func (m *Metadata) ExtensionKind() string { return Kind }

func (m *Metadata) TokenID() string { return m.core.ID() }

func (m *Metadata) TokenURI(tid grc721.TokenID) (string, error) {
	v := m.st.uris.Get(tid.String())
	if v == nil {
		return "", ErrNoTokenURI
	}

	return v.(string), nil
}

func (m *Metadata) TokenMetadata(tid grc721.TokenID) (Data, error) {
	v := m.st.data.Get(tid.String())
	if v == nil {
		return Data{}, ErrNoMetadata
	}

	return v.(Data), nil
}

func (m *Metadata) HasTokenURI(tid grc721.TokenID) bool {
	return m.st.uris.Has(tid.String())
}

func (m *Metadata) HasMetadata(tid grc721.TokenID) bool {
	return m.st.data.Has(tid.String())
}

// SetTokenURI sets the EIP-721 tokenURI of tid; issuer-only so a holder cannot
// rewrite the URI after acquisition.
func (led *Ledger) SetTokenURI(tid grc721.TokenID, uri string) error {
	if _, err := led.core.OwnerOf(tid); err != nil {
		return ErrTokenNotMinted
	}

	led.st.uris.Set(tid.String(), uri)

	chain.Emit(
		TokenURIUpdateEvent,
		"token", led.core.ID(),
		"tokenId", tid.String(),
		"uri", uri,
	)

	return nil
}

// SetTokenMetadata sets the OpenSea on-chain metadata of tid; issuer-only.
func (led *Ledger) SetTokenMetadata(tid grc721.TokenID, data Data) error {
	if _, err := led.core.OwnerOf(tid); err != nil {
		return ErrTokenNotMinted
	}

	led.st.data.Set(tid.String(), data)

	chain.Emit(
		MetadataUpdateEvent,
		"token", led.core.ID(),
		"tokenId", tid.String(),
	)

	return nil
}

func (led *Ledger) ExtensionKind() string { return Kind }

func (led *Ledger) OnMint(to address, tid grc721.TokenID)           {}
func (led *Ledger) OnTransfer(from, to address, tid grc721.TokenID) {}

// OnBurn clears per-token state so it cannot resurface on re-mint.
func (led *Ledger) OnBurn(tid grc721.TokenID) {
	tidStr := tid.String()
	led.st.uris.Remove(tidStr)
	led.st.data.Remove(tidStr)
}
