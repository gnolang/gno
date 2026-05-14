package cla

import (
	"chain"

	"gno.land/p/moul/addrset"
	"gno.land/p/moul/helplink"
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

	desc := "Propose updating the CLA requirement.\n\n"
	if requiredHash != "" {
		desc += "Current hash: " + requiredHash + "\n"
	}
	if claURL != "" {
		desc += "Current URL: " + claURL + "\n"
	}
	if newHash != "" {
		desc += "New hash: " + newHash + "\n"
	}
	if newURL != "" {
		desc += "New URL: " + newURL + "\n"
	}
	if newHash == "" {
		desc += "This proposal disables CLA enforcement.\n"
	}

	return dao.NewProposalRequest(
		"Update CLA requirement",
		desc,
		dao.NewSimpleExecutor(cb, helplink.Realm("gno.land/r/sys/cla").Home()),
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
