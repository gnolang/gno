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
	"chain"
	"chain/runtime"
	"sort"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

const errNoChangesProposed = "no set changes proposed"

// NewProposalRequest builds a GovDAO proposal that, when executed,
// applies the deltas from changesFn to the chain's effective valset
// and publishes the new full set via SetValsetProposal.
//
// changesFn is evaluated EAGERLY at creation; the resulting slice is
// captured into the executor and replayed against the execution-time
// effective baseline. Description is rendered against creation-time
// deltas; application is against execution-time baseline. If a prior
// proposal landed between creation and execution, the captured
// changes apply on top of its result (not on top of pre-creation
// state).
//
// Each Validator with VotingPower > 0 is an add; VotingPower == 0 is
// a removal. Power-update is expressed as the [{power=0},{power=N}]
// remove-then-add pair within one proposal.
func NewProposalRequest(changesFn func() []validators.Validator, title, description string) dao.ProposalRequest {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}

	changes := changesFn()
	if len(changes) == 0 {
		panic(errNoChangesProposed)
	}
	if len(changes) > 40 {
		panic("max number of allowed validators per proposal is 40")
	}

	// For adds, the proposer-supplied Address must equal the
	// pubkey-derived consensus address. Otherwise the proposal
	// description (rendered against c.Address below) would advertise
	// A while the callback's published entry (which encodes only
	// c.PubKey, see line ~129) makes consensus apply pk_B's owner —
	// voters and the on-chain UI would see different identities than
	// what actually lands. Removes (Power==0) carry no PubKey to
	// verify; their c.Address is the lookup key into the effective
	// set and a mismatch panics at execution time.
	for _, c := range changes {
		if c.VotingPower == 0 {
			continue
		}
		derived, err := chain.PubKeyAddress(c.PubKey)
		if err != nil {
			panic("invalid pubkey: " + err.Error())
		}
		if derived != c.Address {
			panic("address/pubkey mismatch: address=" +
				c.Address.String() + " pubkey-derived=" + derived.String())
		}
	}

	var desc strings.Builder
	desc.WriteString(description)
	if len(description) > 0 {
		desc.WriteString("\n\n")
	}
	desc.WriteString("## Validator Updates\n")
	for _, c := range changes {
		if c.VotingPower == 0 {
			desc.WriteString(ufmt.Sprintf("- %s: remove\n", c.Address))
		} else {
			desc.WriteString(ufmt.Sprintf("- %s: add\n", c.Address))
		}
	}

	return dao.NewProposalRequest(title, desc.String(), newValsetChangeExecutor(func() []validators.Validator {
		return changes
	}))
}

// newValsetChangeExecutor builds a GovDAO executor that, on approval,
// applies the captured deltas to a fresh read of the effective valset
// and publishes the result. Unexported by design — only callers
// within this realm should construct executors that reach
// SetValsetProposal.
func newValsetChangeExecutor(changesFn func() []validators.Validator) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	changes := changesFn()
	if len(changes) == 0 {
		panic(errNoChangesProposed)
	}

	callback := func(cur realm) error {
		// Read effective view at execution time: proposed-if-dirty,
		// else current. Robust to:
		//   1. Cross-block sequential — dirty cleared by EndBlocker,
		//      so we read current (== prior proposal's applied result).
		//   2. Same-block sequential — a prior callback wrote
		//      proposed + dirty=true; we accumulate on top instead
		//      of clobbering its result.
		baseline := sysparams.GetValsetEffective()
		set := make(map[address]validators.Validator, len(baseline))
		for _, v := range baseline {
			set[v.Address] = v
		}

		for _, c := range changes {
			if c.VotingPower == 0 {
				if _, ok := set[c.Address]; !ok {
					panic("validator does not exist: " + c.Address.String())
				}
				delete(set, c.Address)
			} else {
				// Add OR power-update. PoA semantics today reject re-add;
				// preserve to keep proposal-author intent unambiguous (use
				// the [{power=0},{power=N}] pair to update).
				if _, ok := set[c.Address]; ok {
					panic("validator already exists: " + c.Address.String())
				}
				set[c.Address] = c
			}
		}

		// Reject all-remove / empty-result proposals before publishing.
		// The chain's EndBlocker rejects an empty proposed set as a
		// liveness floor; without this guard, the proposal would still
		// land here, get published, and silently no-op at apply time
		// while GovDAO already recorded it as accepted (govdao marks
		// acceptance before callback returns). Failing in the callback
		// at least surfaces the no-op as a proposal-execution error.
		if len(set) == 0 {
			panic("valset proposal would empty the validator set; refused to keep consensus liveness")
		}

		// Encode and publish. Sort before publishing — gno map
		// iteration is unspecified; the chain-side diff is set-based
		// so consensus is unaffected, but deterministic bytes keep
		// cross-node diffs clean.
		entries := make([]string, 0, len(set))
		for _, v := range set {
			entries = append(entries, v.PubKey+":"+strconv.FormatUint(v.VotingPower, 10))
		}
		sort.Strings(entries)
		sysparams.SetValsetProposal(cross, entries)
		return nil
	}

	return dao.NewSimpleExecutor(callback, "")
}

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
