package validators

import (
	"sort"
	"strconv"
	"strings"

	"chain"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// ValoperChange is the operator-keyed input shape for the v3 valset
// proposal builder. Power=0 removes; Power>0 adds (or sets power on
// fresh add — re-add of an existing validator is rejected at exec).
//
// Power-update is encoded as the [{op, 0}, {op, N}] remove-then-add
// pair within one proposal — same convention as the legacy
// signing-keyed NewProposalRequest.
type ValoperChange struct {
	OperatorAddress address
	Power           uint64
}

const errNoValoperChanges = "no valoper changes proposed"

// NewValidatorProposalRequest builds a GovDAO proposal that, when
// executed, applies the deltas to the chain's effective valset and
// publishes the new full set via SetValsetProposal.
//
// NON-CROSSING (no `cur realm`). Direct MsgCall is unsupported;
// proposers route through r/gnops/valopers/proposal's facade
// (which IS crossing and accepts user txs). Mirrors the existing
// v3 NewProposalRequest convention.
//
// Validation at creation time: every ValoperChange's OperatorAddress
// must exist in valoperCache (unknown operator panics). This
// enforces validator-must-be-valoper before the proposal goes to a
// vote.
//
// Pubkey resolution at execution time: the executor callback
// re-reads valoperCache for each entry to capture the CURRENT
// signing pubkey/address — not the creation-time one. Defends
// against a stale (now-retired) key publication if the operator
// rotated while the proposal sat in GovDAO. Removes are unaffected
// (operator address is the lookup key, not signing address).
//
// Emits ValidatorAdded / ValidatorRemoved events per entry on
// successful execution.
func NewValidatorProposalRequest(changes []ValoperChange, title, description string) dao.ProposalRequest {
	if len(changes) == 0 {
		panic(errNoValoperChanges)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}
	if len(changes) > 40 {
		panic("max number of allowed validators per proposal is 40")
	}

	// Creation-time validation: every operator must exist in cache.
	// Unknown operators panic now, before voters see the proposal.
	for _, c := range changes {
		if !valoperCache.Has(c.OperatorAddress.String()) {
			panic("unknown operator: " + c.OperatorAddress.String())
		}
	}

	// Render description against creation-time data. Voters see the
	// operator addresses being proposed; signing addresses are an
	// implementation detail resolved at exec.
	var desc strings.Builder
	desc.WriteString(description)
	if len(description) > 0 {
		desc.WriteString("\n\n")
	}
	desc.WriteString("## Validator Updates\n")
	for _, c := range changes {
		if c.Power == 0 {
			desc.WriteString(ufmt.Sprintf("- %s: remove\n", c.OperatorAddress))
		} else {
			desc.WriteString(ufmt.Sprintf("- %s: add (power %d)\n", c.OperatorAddress, c.Power))
		}
	}

	return dao.NewProposalRequest(title, desc.String(), newValoperChangeExecutor(changes))
}

// newValoperChangeExecutor builds the GovDAO executor that, on
// approval, applies the captured ValoperChange deltas. Resolves
// operator → signing addr/pubkey via valoperCache at execution time
// for adds (so a mid-flight rotation doesn't publish a stale key).
// Removes resolve the operator's CURRENT signing address (also via
// cache) — operator-keyed removes are immune to rotation churn.
func newValoperChangeExecutor(changes []ValoperChange) dao.Executor {
	callback := func(cur realm) error {
		baseline := sysparams.GetValsetEffective()
		set := make(map[address]validators.Validator, len(baseline))
		for _, v := range baseline {
			set[v.Address] = v
		}

		for _, c := range changes {
			rawCache, ok := valoperCache.Get(c.OperatorAddress.String())
			if !ok {
				panic("operator vanished from valoperCache between propose and execute: " + c.OperatorAddress.String())
			}
			entry := rawCache.(cacheEntry)

			if c.Power == 0 {
				if _, ok := set[entry.SigningAddress]; !ok {
					panic("validator does not exist: " + entry.SigningAddress.String())
				}
				delete(set, entry.SigningAddress)
				chain.Emit(
					"ValidatorRemoved",
					"op", c.OperatorAddress.String(),
					"signingAddr", entry.SigningAddress.String(),
				)
				continue
			}

			// Add (or fresh-set power). Re-add of an existing entry
			// rejected — use [{op,0},{op,N}] for power update.
			if _, exists := set[entry.SigningAddress]; exists {
				panic("validator already exists: " + entry.SigningAddress.String())
			}
			set[entry.SigningAddress] = validators.Validator{
				Address:     entry.SigningAddress,
				PubKey:      entry.SigningPubKey,
				VotingPower: c.Power,
			}
			chain.Emit(
				"ValidatorAdded",
				"op", c.OperatorAddress.String(),
				"signingAddr", entry.SigningAddress.String(),
				"power", strconv.FormatUint(c.Power, 10),
			)
		}

		// Liveness floor: refuse to publish an empty set.
		if len(set) == 0 {
			panic("valset proposal would empty the validator set; refused to keep consensus liveness")
		}

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
