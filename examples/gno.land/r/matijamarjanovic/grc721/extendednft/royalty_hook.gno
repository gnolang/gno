package extendednft

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/ufmt"
)

type RoyaltyHook struct {
	tokenRoyalties    *avl.Tree // TokenID -> RoyaltyInfo
	maxRoyaltyPercent uint64
}

func NewRoyaltyHook(maxRoyaltyPercent uint64) *RoyaltyHook {
	return &RoyaltyHook{
		tokenRoyalties:    avl.NewTree(),
		maxRoyaltyPercent: maxRoyaltyPercent,
	}
}

func (h *RoyaltyHook) GetHookType() grc721remake.HookType {
	return grc721remake.HookTransfer
}

func (h *RoyaltyHook) GetHookTime() grc721remake.HookTime {
	return grc721remake.HookAfter
}

func (h *RoyaltyHook) OnTransfer(nft *grc721remake.NFT, from, to std.Address, tid TokenID) error {
	if from == std.Address("") || to == std.Address("") {
		return nil
	}

	val, found := h.tokenRoyalties.Get(string(tid))
	if !found {
		return nil
	}

	royaltyInfo := val.(RoyaltyInfo)

	// XXX : this is where we would handle the actual royalty payment
	// For this example, we just emit an event
	std.Emit(
		"RoyaltyPayment",
		"tokenId", string(tid),
		"from", string(from),
		"to", string(to),
		"royaltyReceiver", string(royaltyInfo.PaymentAddress),
		"royaltyPercent", ufmt.Sprintf("%d", royaltyInfo.Percentage),
	)

	return nil
}

func (h *RoyaltyHook) SetTokenRoyalty(tid TokenID, info RoyaltyInfo) error {
	if info.Percentage > h.maxRoyaltyPercent {
		return ufmt.Errorf("royalty percentage exceeds maximum allowed (%d)", h.maxRoyaltyPercent)
	}

	if !info.PaymentAddress.IsValid() {
		return ufmt.Errorf("invalid royalty payment address")
	}

	h.tokenRoyalties.Set(string(tid), info)
	return nil
}

func (h *RoyaltyHook) GetTokenRoyalty(tid TokenID) (RoyaltyInfo, bool) {
	val, found := h.tokenRoyalties.Get(string(tid))
	if !found {
		return RoyaltyInfo{}, false
	}
	return val.(RoyaltyInfo), true
}

func (h *RoyaltyHook) CalculateRoyaltyAmount(tid TokenID, salePrice uint64) (std.Address, uint64, error) {
	royaltyInfo, found := h.GetTokenRoyalty(tid)
	if !found {
		return std.Address(""), 0, ufmt.Errorf("no royalty info for token %s", tid)
	}

	royaltyAmount := (salePrice * royaltyInfo.Percentage) / 100
	return royaltyInfo.PaymentAddress, royaltyAmount, nil
}
