// Package validators implements on-chain validator set management through
// Proof of Authority. The realm exposes a public executor for GovDAO
// proposals to suggest validator set changes. Approved changes are
// published to the VM params keeper via gno.land/r/sys/params; the chain
// (EndBlocker) reads them on the next block and propagates to consensus.
package validators

import (
	"chain/runtime"
	"strings"

	"gno.land/p/nt/poa/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

const errNoChangesProposed = "no set changes proposed"

var vp validators.ValsetProtocol // underlying validator set protocol

func init() {
	vp = poa.NewPoA()
	sysparams.InitValsetPrev()
}

// NewProposalRequest wraps NewValsetChangeExecutor with a GovDAO proposal
// title/description (mirrors v2's NewPropRequest signature).
func NewProposalRequest(changesFn func() []validators.Validator, title, description string) dao.ProposalRequest {
	title = strings.TrimSpace(title)
	if title == "" {
		title = "Valset changes proposal"
	}
	if strings.TrimSpace(description) == "" {
		description = "This proposal applies one or more validator set changes via r/sys/validators/v3."
	}
	return dao.NewProposalRequest(title, description, NewValsetChangeExecutor(changesFn))
}

// NewValsetChangeExecutor creates a GovDAO executor for proposing valset
// changes. changesFn is evaluated eagerly; the resulting slice is captured
// by the returned Executor. Lazy evaluation would close over the caller's
// ephemeral realm and fail to persist via dao.MustCreateProposal.
func NewValsetChangeExecutor(changesFn func() []validators.Validator) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	changes := changesFn()
	if len(changes) == 0 {
		panic(errNoChangesProposed)
	}

	callback := func(cur realm) error {
		// Apply each change to the on-chain valset.
		for _, change := range changes {
			if change.VotingPower == 0 {
				if _, err := vp.RemoveValidator(change.Address); err != nil {
					panic(err)
				}
				continue
			}
			if _, err := vp.AddValidator(change.Address, change.PubKey, change.VotingPower); err != nil {
				panic(err)
			}
		}

		// Publish the new full valset for the chain to diff & apply.
		set := vp.GetValidators()
		pubKeys := make([]string, len(set))
		powers := make([]int64, len(set))
		for i, v := range set {
			pubKeys[i] = v.PubKey.String()
			powers[i] = int64(v.VotingPower)
		}
		sysparams.SetValsetProposal(pubKeys, powers)
		return nil
	}

	return dao.NewSimpleExecutor(callback, "")
}

// IsValidator returns true if addr is part of the current validator set.
func IsValidator(addr address) bool { return vp.IsValidator(addr) }

// GetValidator returns the validator with the given address; panics if absent.
func GetValidator(addr address) validators.Validator {
	v, err := vp.GetValidator(addr)
	if err != nil {
		panic("validator not found")
	}
	return v
}

// GetValidators returns the current validator set.
func GetValidators() []validators.Validator { return vp.GetValidators() }

// Render displays the current validator set.
func Render(string) string {
	var (
		sb  strings.Builder
		h   = runtime.ChainHeight()
		set = GetValidators()
	)
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
