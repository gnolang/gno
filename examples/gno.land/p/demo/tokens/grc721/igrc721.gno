package grc721

type IGRC721 interface {
	BalanceOf(owner address) (int64, error)
	OwnerOf(tid TokenID) (address, error)
	SetTokenURI(_ int, rlm realm, tid TokenID, tURI TokenURI) (bool, error)
	SafeTransferFrom(_ int, rlm realm, from, to address, tid TokenID) error
	TransferFrom(_ int, rlm realm, from, to address, tid TokenID) error
	Approve(_ int, rlm realm, approved address, tid TokenID) error
	SetApprovalForAll(_ int, rlm realm, operator address, approved bool) error
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
)

type NFTGetter func() IGRC721
