package params

const (
	bankModulePrefix    = "bank"
	lockTransferKey     = "lockTransfer"
	unlockTransferTitle = "Proposal to unlock the transfer of ugnot."
	lockTransferTitle   = "Proposal to lock the transfer of ugnot."
)

func ProposeUnlockTransfer() uint64 {
	exe := NewSysParamStringsPropExecutor(bankModulePrefix, "_", lockTransferKey, []string{})
	return propose(exe, unlockTransferTitle, "")
}

func ProposeLockTransfer() uint64 {
	exe := NewSysParamStringsPropExecutor(bankModulePrefix, "_", lockTransferKey, []string{"ugnot"})
	return propose(exe, lockTransferTitle, "")
}
