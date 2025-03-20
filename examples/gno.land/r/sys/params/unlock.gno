package params

import "gno.land/r/gov/dao"

const (
	bankModulePrefix    = "bank"
	restrictedDenomsKey = "restricted_denoms"
	unlockTransferTitle = "Proposal to unlock the transfer of ugnot."
	lockTransferTitle   = "Proposal to lock the transfer of ugnot."
)

func ProposeUnlockTransferRequest() dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(bankModulePrefix, "p", restrictedDenomsKey, unlockTransferTitle, []string{})
}

func ProposeLockTransferRequest() dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(bankModulePrefix, "p", restrictedDenomsKey, lockTransferTitle, []string{"ugnot"})
}
