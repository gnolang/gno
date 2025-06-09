package params

import (
	"gno.land/r/gov/dao"
	"std"
	"sys/params"
)

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

func ProposeAddUnrestrictedAcctsRequest(addrs ...std.Address) dao.ProposalRequest {

	addrStrings := params.GetSysParamStrings(authModulePrefix, "p", unrestrictedAddrsKey)

	// Temporary map for duplicate detection
	existing := make(map[string]struct{}, len(addrStrings))
	for _, s := range addrStrings {
		existing[s] = struct{}{}
	}

	// Append only non-duplicate addresses
	for _, addr := range addrs {
		s := addr.String()
		if _, found := existing[s]; !found {
			addrStrings = append(addrStrings, s)
			existing[s] = struct{}{}
		}
	}

	return NewSysParamStringsPropRequestWithTitle(authModulePrefix, "p", unrestrictedAddrsKey, "Add unrestricted transfer accounts", addrStrings)
}

func ProposeRemoveUnrestrictedAcctsRequest(addrs ...std.Address) dao.ProposalRequest {
	addrStrings := params.GetSysParamStrings(authModulePrefix, "p", unrestrictedAddrsKey)

	removeSet := make(map[string]struct{}, len(addrs))

	for _, addr := range addrs {
		removeSet[addr.String()] = struct{}{}
	}

	result := addrStrings[:0] // reuse original memory
	for _, s := range addrStrings {
		if _, found := removeSet[s]; !found {
			result = append(result, s)
		}
	}

	return NewSysParamStringsPropRequestWithTitle(authModulePrefix, "p", unrestrictedAddrsKey, "Remove unrestricted transfer accounts", result)
}
