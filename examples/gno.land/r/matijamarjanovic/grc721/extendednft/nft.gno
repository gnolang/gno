package extendednft

import (
	"std"

	"gno.land/p/demo/grc/grc721remake"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/authz"
)

var (
	NFT, ledger = grc721remake.NewNFT("Extended NFT", "ENFT")

	Teller = NFT.CallerTeller()

	royaltyHook  = NewRoyaltyHook(100) // Max 100% royalty
	metadataHook = NewMetadataHook()

	auth = authz.New()
)

func init() {
	ledger.RegisterHook("royalty", royaltyHook)
	ledger.RegisterHook("metadata", metadataHook)
}

// Royalty-related functions

func SetTokenRoyalty(tid TokenID, paymentAddress std.Address, percentage uint64) error {
	caller := std.PreviousRealm().Address()

	owner, err := NFT.OwnerOf(tid)
	if err != nil {
		return ufmt.Errorf("token does not exist: %v", err)
	}

	if owner != caller {
		return ufmt.Errorf("caller is not token owner")
	}

	return royaltyHook.SetTokenRoyalty(tid, RoyaltyInfo{
		PaymentAddress: paymentAddress,
		Percentage:     percentage,
	})
}

func GetTokenRoyalty(tid TokenID) (RoyaltyInfo, error) {
	info, found := royaltyHook.GetTokenRoyalty(tid)
	if !found {
		return RoyaltyInfo{}, ufmt.Errorf("no royalty info for token %s", tid)
	}
	return info, nil
}

func CalculateRoyalty(tid TokenID, salePrice uint64) (std.Address, uint64, error) {
	return royaltyHook.CalculateRoyaltyAmount(tid, salePrice)
}

// Metadata-related functions

func SetTokenMetadata(tid TokenID, metadata Metadata) error {
	caller := std.PreviousRealm().Address()
	return metadataHook.SetTokenMetadata(NFT, tid, metadata, caller)
}

func GetTokenMetadata(tid TokenID) (Metadata, error) {
	metadata, found := metadataHook.GetTokenMetadata(tid)
	if !found {
		return Metadata{}, ufmt.Errorf("no metadata for token %s", tid)
	}
	return metadata, nil
}

// Admin functions

func MintToken(to std.Address, tid TokenID) error {
	return auth.Do("mint", func() error {
		return ledger.Mint(to, tid)
	})
}

func BatchMintTokens(to std.Address, tids []TokenID) error {
	return auth.Do("batch_mint", func() error {
		return ledger.BatchMint(to, tids)
	})
}

func BurnToken(tid TokenID) error {
	return auth.Do("burn", func() error {
		return ledger.Burn(tid)
	})
}

func TransferAuthority(newAuthority authz.Authority) error {
	return auth.Transfer(newAuthority)
}

func AddAdmin(addr std.Address) error {
	if memberAuth, ok := auth.Current().(*authz.MemberAuthority); ok {
		return memberAuth.AddMember(addr)
	}
	return ufmt.Errorf("current authority is not a member authority")
}

func RemoveAdmin(addr std.Address) error {
	if memberAuth, ok := auth.Current().(*authz.MemberAuthority); ok {
		return memberAuth.RemoveMember(addr)
	}
	return ufmt.Errorf("current authority is not a member authority")
}

func Render(path string) string {
	if path == "" {
		return renderHome()
	}

	return "Not found"
}

func renderHome() string {
	output := "# " + NFT.GetName() + " (" + NFT.GetSymbol() + ")\n\n"
	output += "An extended NFT collection with royalty and metadata support.\n\n"
	output += "Total supply: " + ufmt.Sprintf("%d", NFT.TokenCount()) + "\n"

	return output
}
