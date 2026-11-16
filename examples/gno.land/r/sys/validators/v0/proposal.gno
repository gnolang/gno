package validators

import (
	"sort"
	"strconv"
	"strings"

	"chain"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sys/validators/v0"
	"gno.land/r/gov/dao"
	sysparams "gno.land/r/sys/params"
)

// ValoperChange is one operator's entry in a valset proposal. Power 0 removes;
// Power > 0 adds, or sets the power of an active operator. Each operator
// appears at most once per proposal.
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

// NewValidatorProposalRequest builds a GovDAO proposal applying changes to the
// effective valset through SetValsetProposal. Proposers reach it through
// r/gnops/valopers/proposal.
//
// At creation every operator must be known to valoperCache, appear once, and
// have KeepRunning set if added. At execution the signing key and KeepRunning
// are read again, so a rotation or opt-out during the vote is honored.
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

	// Snapshot: the caller can still write the array it passed. Copied before
	// any check, so what is checked, described and executed is one array.
	owned := make([]ValoperChange, len(changes))
	copy(owned, changes)
	changes = owned

	// Each operator at most once; a power change is one upsert entry.
	seen := map[string]bool{}
	for _, c := range changes {
		key := c.OperatorAddress.String()
		if seen[key] {
			panic("duplicate operator in proposal: " + key)
		}
		seen[key] = true
	}

	// Every operator must be known, and an add needs KeepRunning.
	for _, c := range changes {
		rawCache := valoperCache.Get(c.OperatorAddress.String())
		if rawCache == nil {
			panic("unknown operator: " + c.OperatorAddress.String())
		}
		entry := rawCache.(cacheEntry)
		if c.Power > 0 && !entry.KeepRunning {
			panic("operator " + c.OperatorAddress.String() + " has KeepRunning=false; refusing to add (operator must call UpdateKeepRunning(true) first)")
		}
	}

	// Voters see operator addresses; signing keys resolve at execution.
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

	return dao.NewProposalRequest(title, desc.String(), newValoperChangeExecutor(cur, changes))
}

// newValoperChangeExecutor applies changes on approval, resolving each
// operator's current signing key from valoperCache. Power > 0 upserts.
func newValoperChangeExecutor(cur realm, changes []ValoperChange) dao.Executor {
	callback := func(cur realm) error {
		baseline := sysparams.GetValsetEffective()
		set := make(map[address]validators.Validator, len(baseline))
		for _, v := range baseline {
			set[v.Address] = v
		}

		for _, c := range changes {
			rawCache := valoperCache.Get(c.OperatorAddress.String())
			if rawCache == nil {
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

			// KeepRunning may have been turned off during the vote.
			if !entry.KeepRunning {
				panic("operator " + c.OperatorAddress.String() + " has KeepRunning=false at execution; refusing to add")
			}

			// Upsert at the current signing address.
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
		sysparams.SetValsetProposal(cross(cur), entries)
		return nil
	}

	return dao.NewSimpleExecutor(0, cur, callback, "")
}
