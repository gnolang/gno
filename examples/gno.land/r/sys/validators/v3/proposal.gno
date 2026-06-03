package validators

import (
	"math/overflow"
	"sort"
	"strconv"
	"strings"
	"time"

	"chain"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// ValoperChange is the operator-keyed input shape for the v3 valset
// proposal builder. Power=0 removes; Power>0 adds (or upserts the
// power on an op already in the active set — Tendermint's natural
// ValidatorUpdate semantics).
//
// Each operator may appear AT MOST ONCE per proposal; duplicates are
// rejected at create-time.
type ValoperChange struct {
	OperatorAddress address
	Power           uint64
}

func NewValoperChange(operatorAddress address, power uint64) ValoperChange {
	return ValoperChange{
		OperatorAddress: operatorAddress,
		Power:           power,
	}
}

const errNoValoperChanges = "no valoper changes proposed"

// NewValidatorProposalRequest builds a GovDAO proposal that, when
// executed, applies the deltas to the chain's effective valset and
// publishes the new full set via SetValsetProposal.
//
// NON-CROSSING (no `cur realm`). Direct MsgCall is unsupported;
// proposers route through r/gnops/valopers/proposal's facade
// (which IS crossing and accepts user txs).
//
// Validation at creation time:
//   - Each operator may appear AT MOST ONCE in changes; duplicates
//     panic. Power changes for an op already in the active set use
//     a single {op, newPower} entry (upsert), not the legacy
//     remove/re-add pair.
//   - Every ValoperChange's OperatorAddress must exist in
//     valoperCache. Unknown operators panic.
//   - Adds (Power > 0) require KeepRunning=true. An op that has
//     called UpdateKeepRunning(false) signals opt-out; no proposal
//     can keep them in the active set, period.
//
// Pubkey resolution at execution time: the executor callback
// re-reads valoperCache for each entry to capture the CURRENT
// signing pubkey/address — not the creation-time one. Defends
// against a stale (now-retired) key publication if the operator
// rotated while the proposal sat in GovDAO. Also re-checks
// KeepRunning so an operator flipping to KeepRunning=false between
// propose-create and propose-execute is honored. Removes are
// unaffected (operator address is the lookup key, not signing
// address).
//
// Emits ValidatorAdded / ValidatorRemoved events per entry on
// successful execution. (Power-upsert on an existing op also emits
// ValidatorAdded with the new power.)
func NewValidatorProposalRequest(cur realm, changes []ValoperChange, title, description string) dao.ProposalRequest {
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

	// Cooldown: refuse to even queue a proposal if the previous valset
	// update happened too recently. The executor re-checks this — see
	// the comment on lastValsetUpdate in limits.gno.
	if time.Since(lastValsetUpdate) < valsetUpdateCooldown {
		panic(errValsetUpdateCooldown)
	}

	// Dedupe: each operator may appear at most once per proposal.
	// Power changes are now expressed as a single {op, newPower}
	// upsert entry, so the legacy [{op,0},{op,N}] pair is a duplicate
	// and rejected.
	seen := map[string]bool{}
	for _, c := range changes {
		key := c.OperatorAddress.String()
		if seen[key] {
			panic("duplicate operator in proposal: " + key)
		}
		seen[key] = true
	}

	// Creation-time validation: every operator must exist in cache,
	// and adds require KeepRunning=true. KeepRunning=false is a
	// binding opt-out; no proposal shape can override it.
	for _, c := range changes {
		rawCache, ok := valoperCache.Get(c.OperatorAddress.String())
		if !ok {
			panic("unknown operator: " + c.OperatorAddress.String())
		}
		entry := rawCache.(cacheEntry)
		if c.Power > 0 && !entry.KeepRunning {
			panic("operator " + c.OperatorAddress.String() + " has KeepRunning=false; refusing to add (operator must call UpdateKeepRunning(true) first)")
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

	// Snapshot the trust-level ratio at proposal-creation time so the
	// rule a proposal was screened against can't be relaxed between
	// propose and execute. Without this snapshot a same-block sequence
	// "trust-level-drop executor → valset-change executor" would let
	// the second proposal pass under the looser ratio.
	return dao.NewProposalRequest(title, desc.String(), newValoperChangeExecutor(cur, changes, trustLevelRatio))
}

// newValoperChangeExecutor builds the GovDAO executor that, on
// approval, applies the captured ValoperChange deltas. Resolves
// operator → signing addr/pubkey via valoperCache at execution time
// for adds (so a mid-flight rotation doesn't publish a stale key).
// Removes resolve the operator's CURRENT signing address (also via
// cache) — operator-keyed removes are immune to rotation churn.
//
// Power>0 is an upsert against the effective valset map (keyed on
// signing address): if the op is already present under that signing
// address, the entry's voting power is overwritten. Tendermint
// natively handles ValidatorUpdates as upserts, so a single-entry
// power change is the canonical form.
//
// snapshotTrustLevel captures trustLevelRatio at proposal-creation
// time. The IBC-trust-level rule is checked against this snapshot, not
// the package-level trustLevelRatio that may have moved by exec time.
func newValoperChangeExecutor(cur realm, changes []ValoperChange, snapshotTrustLevel trustRatio) dao.Executor {
	callback := func(cur realm) error {
		// Cooldown re-check: the binding source of truth. Creation-time
		// only filters obvious failures; a proposal can sit in GovDAO
		// while another proposal lands first.
		if time.Since(lastValsetUpdate) < valsetUpdateCooldown {
			panic(errValsetUpdateCooldown)
		}

		baseline := sysparams.GetValsetEffective()
		set := make(map[address]validators.Validator, len(baseline))
		// baselineByAddr captures the pre-update voting power keyed by
		// signing address — used by the trust-level check below to
		// measure how much of the previous set survived.
		baselineByAddr := make(map[address]uint64, len(baseline))
		var baselineTotal uint64
		for _, v := range baseline {
			set[v.Address] = v
			baselineByAddr[v.Address] = v.VotingPower
			baselineTotal += v.VotingPower
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

			// Race-safety: operator may have flipped KeepRunning=false
			// between propose-create and propose-execute. Re-check.
			// The opt-out is binding regardless of proposal shape.
			if !entry.KeepRunning {
				panic("operator " + c.OperatorAddress.String() + " has KeepRunning=false at execution; refusing to add")
			}

			// Upsert at the current signing address. If the entry was
			// already present (single-entry power change on an active
			// validator), this overwrites the prior power.
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

		// Trust-level check (IBC light-client rule): the previous
		// validator set must retain at least trustLevelRatio of its own
		// voting power in the new set, so a light client at the old
		// header can verify the new header.
		//
		// "Retained" is weighted by each survivor's BASELINE voting
		// power, not the new one. This matches CometBFT's tally in
		// types/validation.go:verifyCommitSingle — `talliedVotingPower
		// += val.VotingPower` where `val` comes from the TRUSTED
		// validator set, regardless of any new VP the same validator
		// has in the untrusted header. Using new-VP here would create
		// a false-positive accept: a proposal can remove most of the
		// baseline and inflate a single survivor's VP to mask the
		// loss; the light client (using baseline VPs) would refuse to
		// verify the resulting header.
		//
		// Cross-multiply with overflow.Mulu64 so the comparison stays
		// integer-deterministic.
		var retainedFromBaseline uint64
		for addr, baselineVP := range baselineByAddr {
			if _, stillIn := set[addr]; stillIn {
				retainedFromBaseline += baselineVP
			}
		}
		if baselineTotal > 0 {
			left, okL := overflow.Mulu64(retainedFromBaseline, snapshotTrustLevel.denominator)
			right, okR := overflow.Mulu64(baselineTotal, snapshotTrustLevel.numerator)
			if !okL || !okR {
				panic(errTrustLevelOverflow)
			}
			// Strict inequality matches CometBFT's light-client check
			// (types/validation.go:355,503): fail iff
			// talliedVotingPower <= votingPowerNeeded. The boundary
			// case (tallied * den == total * num) is REJECTED here so
			// no chain-side-accepted update can leave an IBC light
			// client unable to verify the next header.
			if left <= right {
				panic(errTrustLevelViolated)
			}
		}

		entries := make([]string, 0, len(set))
		for _, v := range set {
			entries = append(entries, v.PubKey+":"+strconv.FormatUint(v.VotingPower, 10))
		}
		sort.Strings(entries)
		sysparams.SetValsetProposal(cross(cur), entries)
		lastValsetUpdate = time.Now()
		return nil
	}

	return dao.NewSimpleExecutor(0, cur, callback, "")
}
