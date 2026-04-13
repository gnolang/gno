package grc721

type IGRC721 interface {
	BalanceOf(owner address) (int64, error)
	OwnerOf(tid TokenID) (address, error)
	SetTokenURI(tid TokenID, tURI TokenURI) (bool, error)
	SafeTransferFrom(from, to address, tid TokenID) error
	TransferFrom(from, to address, tid TokenID) error
	Approve(approved address, tid TokenID) error
	SetApprovalForAll(operator address, approved bool) error
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
