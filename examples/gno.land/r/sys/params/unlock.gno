package params

const (
	lockSendKey     = "bank_lockSend.string"
	UnlockSendTitle = "Proposal to unlock the sending functionality for ugnot."
	LockSendTitle   = "Proposal to lock the sending functionality for ugnot."
)

func ProposeUnlockSend() uint64 {
	exe := NewStringPropExecutor(lockSendKey, "")
	return propose(exe, UnlockSendTitle, "")
}

func ProposeLockSend() uint64 {
	exe := NewStringPropExecutor(lockSendKey, "ugnot")
	return propose(exe, LockSendTitle, "")
}
