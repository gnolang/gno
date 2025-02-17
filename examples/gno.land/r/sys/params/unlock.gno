package params

const (
	bankKeeperPrefix    = "bank"
	lockTransferKey     = "lockTransfer.string"
	unlockTransferTitle = "Proposal to unlock the transfer of ugnot."
	lockTransferTitle   = "Proposal to lock the transfer of ugnot."
)

func ProposeUnlockTransfer() uint64 {
	exe := NewPrefixedStringPropExecutor(bankKeeperPrefix, lockTransferKey, "")
	return propose(exe, unlockTransferTitle, "")
}

func ProposeLockTransfer() uint64 {
	exe := NewPrefixedStringPropExecutor(bankKeeperPrefix, lockTransferKey, "ugnot")
	return propose(exe, lockTransferTitle, "")
}
