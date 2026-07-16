package grc1155

type IGRC1155 interface {
	SafeTransferFrom(from, to address, tid TokenID, amount int64) error
	SafeBatchTransferFrom(from, to address, batch []TokenID, amounts []int64) error
	BalanceOf(owner address, tid TokenID) (int64, error)
	BalanceOfBatch(owners []address, batch []TokenID) ([]int64, error)
	SetApprovalForAll(operator address, approved bool) error
	IsApprovedForAll(owner, operator address) bool
}

type TokenID string

type TransferSingleEvent struct {
	Operator address
	From     address
	To       address
	TokenID  TokenID
	Amount   int64
}

type TransferBatchEvent struct {
	Operator address
	From     address
	To       address
	Batch    []TokenID
	Amounts  []int64
}

type ApprovalForAllEvent struct {
	Owner    address
	Operator address
	Approved bool
}

type UpdateURIEvent struct {
	URI string
}

type MultiTokenGetter func() IGRC1155
