package extendednft

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/ufmt"
)

type MetadataHook struct {
	tokenMetadata *avl.Tree // TokenID -> Metadata
}

func NewMetadataHook() *MetadataHook {
	return &MetadataHook{
		tokenMetadata: avl.NewTree(),
	}
}

func (h *MetadataHook) GetHookType() grc721remake.HookType {
	return grc721remake.HookUniversal
}

func (h *MetadataHook) GetHookTime() grc721remake.HookTime {
	return grc721remake.HookAfter
}

func (h *MetadataHook) OnAny(nft *grc721remake.NFT, operation string) error {
	// This hook doesn't need to do anything on operations
	// It just stores metadata that can be retrieved later
	return nil
}

func (h *MetadataHook) SetTokenMetadata(nft *grc721remake.NFT, tid TokenID, metadata Metadata, caller std.Address) error {
	owner, err := nft.OwnerOf(tid)
	if err != nil {
		return ufmt.Errorf("token does not exist: %v", err)
	}

	if owner != caller {
		return ufmt.Errorf("caller is not token owner")
	}

	h.tokenMetadata.Set(string(tid), metadata)

	std.Emit(
		"MetadataUpdated",
		"tokenId", string(tid),
		"owner", string(owner),
	)

	return nil
}

func (h *MetadataHook) GetTokenMetadata(tid TokenID) (Metadata, bool) {
	val, found := h.tokenMetadata.Get(string(tid))
	if !found {
		return Metadata{}, false
	}
	return val.(Metadata), true
}
