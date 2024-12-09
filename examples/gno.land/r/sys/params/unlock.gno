package params

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
)

const (
	lockSendKey    = "lockSend.bool"
	UnlockSendDesc = "Proposal to unlock the sending functionality for ugnot."
	LockSendDesc   = "Proposal to lock the sending functionality for ugnot."
)

func ProposeUnlockSend() uint64 {
	callback := func() error {
		std.SetParamBool(lockSendKey, false)
		return nil
	}
	return propose(callback, UnlockSendDesc)
}

func ProposeLockSend() uint64 {
	callback := func() error {
		std.SetParamBool(lockSendKey, true)
		return nil
	}
	return propose(callback, LockSendDesc)
}

func propose(callback func() error, desc string) uint64 {
	// The callback function is executed only after the proposal is voted on
	// and approved by the GovDAO.
	exe := bridge.GovDAO().NewGovDAOExecutor(callback)
	prop := dao.ProposalRequest{
		Description: desc,
		Executor:    exe,
	}
	return bridge.GovDAO().Propose(prop)
}
