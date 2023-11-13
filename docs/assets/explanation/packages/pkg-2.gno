// functions that work similarly to those of grc20
func BalanceOf(owner std.Address) (uint64, error)
func Approve(approved std.Address, tid TokenID) error
func TransferFrom(from, to std.Address, tid TokenID) error

// functions unique to grc721
func OwnerOf(tid TokenID) (std.Address, error)
func SafeTransferFrom(from, to std.Address, tid TokenID) error
func SetApprovalForAll(operator std.Address, approved bool) error
func GetApproved(tid TokenID) (std.Address, error)
func IsApprovedForAll(owner, operator std.Address) bool
