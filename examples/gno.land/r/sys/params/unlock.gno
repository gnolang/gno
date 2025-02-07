package params

const (
	bankKeeperPrefix    = "bank"
	lockTransferKey     = "lockTransfer.string"
	UnlockTransferTitle = "Proposal to unlock the transfer of ugnot."
	LockTransferTitle   = "Proposal to lock the transfer of ugnot."
)

func ProposeUnlockTransfer() uint64 {
	exe := NewPrefixedStringPropExecutor(bankKeeperPrefix, lockTransferKey, "")
	return propose(exe, UnlockTransferTitle, "")
}

func ProposeLockTransfer() uint64 {
	exe := NewPrefixedStringPropExecutor(bankKeeperPrefix, lockTransferKey, "ugnot")
	return propose(exe, LockTransferTitle, "")
}
