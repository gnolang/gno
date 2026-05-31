// Package validators implements on-chain validator set management
// through Proof of Authority. The realm exposes a public proposal
// constructor for GovDAO; on approval, the proposal callback applies
// the captured deltas to the chain's effective valset and publishes
// the new full set via gno.land/r/sys/params. The chain's EndBlocker
// reads the result on the next block and propagates to consensus.
//
// No in-realm validator state. All reads go through
// sysparams.GetValsetEffective (proposed-if-dirty, else current).
package validators

import (
	"chain/runtime"
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	sysparams "gno.land/r/sys/params"
)

// Operator-keyed proposal builder lives in proposal.gno
// (NewValidatorProposalRequest + newValoperChangeExecutor). The legacy
// signing-keyed NewProposalRequest was removed: every valid
// signing-keyed input is also a valid operator-keyed input under
// always-on valoper enforcement.

// IsValidator returns true if addr is part of the effective validator
// set (proposed if a v3 proposal is awaiting EndBlocker, else
// current).
func IsValidator(addr address) bool {
	for _, v := range sysparams.GetValsetEffective() {
		if v.Address == addr {
			return true
		}
	}
	return false
}

// GetValidator returns the validator with the given address from the
// effective set; panics if absent.
func GetValidator(addr address) validators.Validator {
	for _, v := range sysparams.GetValsetEffective() {
		if v.Address == addr {
			return v
		}
	}
	panic("validator not found")
}

// GetValidators returns the effective validator set.
func GetValidators() []validators.Validator {
	return sysparams.GetValsetEffective()
}

// Render displays the effective validator set.
func Render(string) string {
	var sb strings.Builder
	h := runtime.ChainHeight()
	set := sysparams.GetValsetEffective()
	sb.WriteString(ufmt.Sprintf("## Valset at height %d\n\n", h))
	if len(set) == 0 {
		sb.WriteString("Valset is empty.\n")
		return sb.String()
	}
	for i, v := range set {
		sb.WriteString(ufmt.Sprintf("- #%d: %s (%d)\n", i, v.Address.String(), v.VotingPower))
	}
	return sb.String()
}
