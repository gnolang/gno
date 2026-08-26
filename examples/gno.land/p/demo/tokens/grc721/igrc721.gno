package grc721

// IGRC721Reader is the read-only view of an NFT. Safe to receive across
// realm boundaries — has no rlm-typed methods, so a malicious impl can
// only lie about read results (data-integrity issue), not capture cur.
//
// Writes are concrete methods on *BasicNFT / *metadataNFT / *royaltyNFT
// only — there is no IGRC721 writer interface. The owning realm should
// hold the concrete *BasicNFT in an unexported package var and expose
// public rlm-validating wrappers like:
//
//	func TransferFrom(cur realm, from, to address, tid TokenID) error {
//		caller := cur.Previous().Address()
//		return nft.TransferFrom(caller, from, to, tid)
//	}
//
// This is the Reader/Writer split — stronger than the Authority-pattern
// because the writer interface doesn't exist at all, so no realm author
// can accidentally expose it. See `r/demo/foo721` for the canonical
// wrapper pattern.
type IGRC721Reader interface {
	Name() string
	Symbol() string
	TokenCount() int64
	BalanceOf(owner address) (int64, error)
	OwnerOf(tid TokenID) (address, error)
	GetApproved(tid TokenID) (address, error)
	IsApprovedForAll(owner, operator address) bool
}

type (
	TokenID  string
	TokenURI string
)

func (t TokenID) String() string  { return string(t) }
func (t TokenURI) String() string { return string(t) }

const (
	MintEvent           = "Mint"
	BurnEvent           = "Burn"
	TransferEvent       = "Transfer"
	ApprovalEvent       = "Approval"
	ApprovalForAllEvent = "ApprovalForAll"
	TokenURIUpdateEvent = "TokenUriUpdate"
	MetadataUpdateEvent = "MetadataUpdate"
)

// NFTGetter returns a reader-only view of an NFT. Aggregators (such as
// tokenhub) register and dispatch NFTGetters; the reader-only return
// type means even a malicious aggregator can't be used to leak cur.
type NFTGetter func() IGRC721Reader
