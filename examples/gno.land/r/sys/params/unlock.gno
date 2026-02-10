package params

import "gno.land/r/gov/dao"

const (
	bankModulePrefix     = "bank"
	restrictedDenomsKey  = "restricted_denoms"
	unlockTransferTitle  = "Proposal to unlock the transfer of ugnot."
	lockTransferTitle    = "Proposal to lock the transfer of ugnot."
	authModulePrefix     = "auth"
	unrestrictedAddrsKey = "unrestricted_addrs"
)

func ProposeUnlockTransferRequest() dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(bankModulePrefix, "p", restrictedDenomsKey, unlockTransferTitle, []string{})
}

func ProposeLockTransferRequest() dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(bankModulePrefix, "p", restrictedDenomsKey, lockTransferTitle, []string{"ugnot"})
}

func ProposeAddUnrestrictedAcctsRequest(addrs ...address) dao.ProposalRequest {
	addrStrings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		s := addr.String()
		addrStrings = append(addrStrings, s)
	}
	return NewSysParamStringsPropRequestAddWithTitle(authModulePrefix, "p", unrestrictedAddrsKey, "Add unrestricted transfer accounts", addrStrings)
}

func ProposeRemoveUnrestrictedAcctsRequest(addrs ...address) dao.ProposalRequest {
	addrStrings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		s := addr.String()
		addrStrings = append(addrStrings, s)
	}
	return NewSysParamStringsPropRequestRemoveWithTitle(authModulePrefix, "p", unrestrictedAddrsKey, "Add unrestricted transfer accounts", addrStrings)
}
