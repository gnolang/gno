package params

import "gno.land/r/gov/dao"

const (
	bankModulePrefix    = "bank"
	restrictedDenomsKey = "restricted_denoms"
	unlockTransferTitle = "Proposal to unlock the transfer of ugnot."
	lockTransferTitle   = "Proposal to lock the transfer of ugnot."
)

func ProposeUnlockTransfer() dao.ProposalID {
	req := NewSysParamStringPropRequest(bankModulePrefix, "p", restrictedDenomsKey, "")
	return propose(req)
}

func ProposeLockTransfer() dao.ProposalID {
	req := NewSysParamStringPropRequest(bankModulePrefix, "p", restrictedDenomsKey, "ugnot")
	return propose(req)
}
