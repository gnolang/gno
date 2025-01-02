package params

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
)

const (
	lockSendKey     = "lockSend.bool"
	UnlockSendTitle = "Proposal to unlock the sending functionality for ugnot."
	LockSendTitle   = "Proposal to lock the sending functionality for ugnot."
)

func ProposeUnlockSend() uint64 {
	callback := func() error {
		std.SetParamBool(lockSendKey, false)
		return nil
	}
	return propose(callback, UnlockSendTitle, "")
}

func ProposeLockSend() uint64 {
	callback := func() error {
		std.SetParamBool(lockSendKey, true)
		return nil
	}
	return propose(callback, LockSendTitle, "")
}

func propose(callback func() error, title, desc string) uint64 {
	// The callback function is executed only after the proposal is voted on
	// and approved by the GovDAO.
	exe := bridge.GovDAO().NewGovDAOExecutor(callback)
	prop := dao.ProposalRequest{
		Title:       title,
		Description: desc,
		Executor:    exe,
	}
	return bridge.GovDAO().Propose(prop)
}
