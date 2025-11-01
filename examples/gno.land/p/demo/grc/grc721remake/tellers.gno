package grc721remake

import (
	"std"
)

// CallerTeller returns a GRC721 compatible teller that checks the PreviousRealm
// caller for each call. It's usually safe to expose it publicly to let users
// manage their NFTs directly, or for realms to use their approvals.
func (nft *NFT) CallerTeller() NFTTeller {
	if nft == nil {
		panic("NFT cannot be nil")
	}

	return &fnNFTTeller{
		accountFn: func() std.Address {
			caller := std.PreviousRealm().Address()
			return caller
		},
		NFT: nft,
	}
}

// ReadonlyTeller is a GRC721 compatible teller that panics for any write operation.
func (nft *NFT) ReadonlyTeller() NFTTeller {
	if nft == nil {
		panic("NFT cannot be nil")
	}

	return &fnNFTTeller{
		accountFn: nil,
		NFT:       nft,
	}
}

// RealmTeller returns a GRC721 compatible teller that will store the
// caller realm permanently. Calling anything through this teller will
// result in approval or ownership changes for the realm that initialized the teller.
// The initializer of this teller should usually never share the resulting NFTTeller from
// this method except maybe for advanced delegation flows such as a DAO treasury
// management.
func (nft *NFT) RealmTeller() NFTTeller {
	if nft == nil {
		panic("NFT cannot be nil")
	}

	caller := std.PreviousRealm().Address()

	return &fnNFTTeller{
		accountFn: func() std.Address {
			return caller
		},
		NFT: nft,
	}
}

// RealmSubTeller is like RealmTeller but uses the provided slug to derive a
// subaccount.
func (nft *NFT) RealmSubTeller(slug string) NFTTeller {
	if nft == nil {
		panic("NFT cannot be nil")
	}

	caller := std.PreviousRealm().Address()
	account := accountSlugAddr(caller, slug)

	return &fnNFTTeller{
		accountFn: func() std.Address {
			return account
		},
		NFT: nft,
	}
}

// ImpersonateTeller returns a GRC721 compatible teller that impersonates as a
// specified address. This allows operations to be performed as if they were
// executed by the given address, enabling the caller to manipulate NFTs on
// behalf of that address.
//
// It is particularly useful in scenarios where a contract needs to perform
// actions on behalf of a user or another account, without exposing the
// underlying logic or requiring direct access to the user's account. The
// returned teller will use the provided address for all operations, effectively
// masking the original caller.
//
// This method should be used with caution, as it allows for potentially
// sensitive operations to be performed under the guise of another address.
func (ledger *PrivateLedger) ImpersonateTeller(addr std.Address) NFTTeller {
	if ledger == nil {
		panic("Ledger cannot be nil")
	}

	return &fnNFTTeller{
		accountFn: func() std.Address {
			return addr
		},
		NFT: ledger.nft,
	}
}

// Generic tellers methods implementation

func (ft *fnNFTTeller) SetTokenURI(tid TokenID, tURI TokenURI) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.SetTokenURI(caller, tid, tURI)
}
func (ft *fnNFTTeller) Approve(to std.Address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.Approve(caller, to, tid)
}

func (ft *fnNFTTeller) SetApprovalForAll(operator std.Address, approved bool) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.SetApprovalForAll(caller, operator, approved)
}

func (ft *fnNFTTeller) Transfer(to std.Address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.Transfer(caller, to, tid)
}

func (ft *fnNFTTeller) TransferFrom(from, to std.Address, tid TokenID) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.TransferFrom(caller, from, to, tid)
}

func (ft *fnNFTTeller) BatchTransfer(to std.Address, tids []TokenID) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.BatchTransfer(caller, to, tids)
}

func (ft *fnNFTTeller) BatchTransferFrom(from, to std.Address, tids []TokenID) error {
	if ft.accountFn == nil {
		return ErrReadOnly
	}

	caller := ft.accountFn()
	return ft.NFT.ledger.BatchTransferFrom(caller, from, to, tids)
}

// helpers

// accountSlugAddr returns the address derived from the specified address and slug.
func accountSlugAddr(addr std.Address, slug string) std.Address {
	// XXX: use a new `std.XXX` call for this.
	if slug == "" {
		return addr
	}
	key := addr.String() + "/" + slug
	return std.DerivePkgAddr(key) // temporarily using this helper
}
