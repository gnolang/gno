package cla

import (
	"chain"

	"gno.land/p/moul/addrset"
	"gno.land/r/gov/dao"
)

const RequiredHashChangedEvent = "CLARequiredHashChanged"

// ProposeNewCLA creates a govdao proposal to update the CLA document hash and URL.
// When executed, it resets all existing signatures.
// Propose an empty hash to disable CLA enforcement.
func ProposeNewCLA(newHash, newURL string) dao.ProposalRequest {
	cb := func(cur realm) error {
		setRequiredHash(newHash)
		claURL = newURL
		return nil
	}

	return dao.NewProposalRequest(
		"Update CLA requirement",
		"",
		dao.NewSimpleExecutor(cb, ""),
	)
}

func setRequiredHash(newHash string) {
	prevHash := requiredHash
	requiredHash = newHash
	signatures = addrset.Set{} // reset all signatures

	chain.Emit(
		RequiredHashChangedEvent,
		"from", prevHash,
		"to", newHash,
	)
}
