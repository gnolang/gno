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

func ProposeUnlockTransferRequest(cur realm) dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(cur, bankModulePrefix, "p", restrictedDenomsKey, unlockTransferTitle, []string{})
}

func ProposeLockTransferRequest(cur realm) dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(cur, bankModulePrefix, "p", restrictedDenomsKey, lockTransferTitle, []string{"ugnot"})
}

func ProposeAddUnrestrictedAcctsRequest(cur realm, addrs ...address) dao.ProposalRequest {
	addrStrings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		s := addr.String()
		addrStrings = append(addrStrings, s)
	}
	return NewSysParamStringsPropRequestAddWithTitle(cur, authModulePrefix, "p", unrestrictedAddrsKey, "Add unrestricted transfer accounts", addrStrings)
}

func ProposeRemoveUnrestrictedAcctsRequest(cur realm, addrs ...address) dao.ProposalRequest {
	addrStrings := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		s := addr.String()
		addrStrings = append(addrStrings, s)
	}
	return NewSysParamStringsPropRequestRemoveWithTitle(cur, authModulePrefix, "p", unrestrictedAddrsKey, "Add unrestricted transfer accounts", addrStrings)
}
