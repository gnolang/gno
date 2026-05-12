package validators

import (
	"errors"
	"math/overflow"
	"strings"
	"time"

	"gno.land/r/gov/dao"
)

// Valset-update safety limits (IBC trust-level + cooldown).
//
//   - lastValsetUpdate: timestamp of the most recent successful valset
//     proposal execution. Read on both proposal create and execute to
//     enforce the cooldown; written only after a proposal's executor
//     finishes a successful SetValsetProposal call.
//
//   - valsetUpdateCooldown: minimum elapsed time between two consecutive
//     valset updates. The acceptance criteria in #4829 specify 24h.
//
//   - trustLevelRatio: the fraction of the previous voting power that the
//     surviving previous validators must retain in the new set. The IBC
//     light-client spec bounds this to [1/3, 2/3] (1/3 is the BFT honest
//     floor; >2/3 would defeat the whole purpose). Update via the
//     NewTrustLevelPropRequest governance proposal.
var (
	lastValsetUpdate     time.Time
	valsetUpdateCooldown = 24 * time.Hour
	trustLevelRatio      = trustRatio{numerator: 1, denominator: 3}
	trustLevelMinAllowed = trustRatio{numerator: 1, denominator: 3}
	trustLevelMaxAllowed = trustRatio{numerator: 2, denominator: 3}
)

const (
	errValsetUpdateCooldown = "valset update cooldown in effect"
	errTrustLevelViolated   = "trust level violated: insufficient baseline voting power retained"
	errInvalidTrustLevel    = "invalid trust level fraction"
	errTrustLevelOverflow   = "trust level arithmetic overflow"
	errInvalidCooldown      = "invalid cooldown"

	// cooldownMaxSeconds caps NewCooldownPropRequest input at one year.
	// time.Duration is int64 nanoseconds, so unbounded seconds * 1e9
	// silently overflows past ~9.22e9 seconds (≈292 years) and could
	// wrap to a tiny or negative duration — effectively disabling the
	// cooldown by accident. A 1-year ceiling is far beyond any
	// reasonable governance choice and well inside int64 range.
	cooldownMaxSeconds uint64 = 365 * 24 * 60 * 60
)

// trustRatio represents num/den. Comparisons are done by cross-multiplication
// to stay deterministic across architectures (no floats).
type trustRatio struct {
	numerator   uint64
	denominator uint64
}

// lessThan reports whether tr < other. Cross-multiply: a/b < c/d iff a*d < c*b.
// Uses overflow-checked multiplication; an overflow on either side panics
// rather than returning a silently-wrong comparison. Denominators are
// non-zero by construction (NewTrustLevelPropRequest rejects denominator==0
// before calling this; trustLevelMin/Max have denominator=3).
func (tr trustRatio) lessThan(other trustRatio) bool {
	left, okL := overflow.Mulu64(tr.numerator, other.denominator)
	right, okR := overflow.Mulu64(other.numerator, tr.denominator)
	if !okL || !okR {
		panic(errTrustLevelOverflow)
	}
	return left < right
}

// GetTrustLevel returns the currently configured trust level as (numerator, denominator).
func GetTrustLevel() (uint64, uint64) {
	return trustLevelRatio.numerator, trustLevelRatio.denominator
}

// NewTrustLevelPropRequest builds a GovDAO proposal that, when executed,
// updates trustLevelRatio. The ratio must be in [1/3, 2/3] both at
// creation and execution time (re-checked in the callback in case bounds
// change between propose and execute).
func NewTrustLevelPropRequest(numerator, denominator uint64, title, description string) dao.ProposalRequest {
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}
	if denominator == 0 {
		panic(errInvalidTrustLevel)
	}
	newRatio := trustRatio{numerator: numerator, denominator: denominator}
	if newRatio.lessThan(trustLevelMinAllowed) || trustLevelMaxAllowed.lessThan(newRatio) {
		panic(errInvalidTrustLevel)
	}

	return dao.NewProposalRequest(title, description, newTrustLevelExecutor(newRatio))
}

// newTrustLevelExecutor builds the GovDAO executor that applies a
// trust-level update. The bounds are re-checked at execute-time in case
// trustLevelMinAllowed / trustLevelMaxAllowed changed between propose
// and execute (today they're constants, but the executor stays defensive).
func newTrustLevelExecutor(newRatio trustRatio) dao.Executor {
	callback := func(cur realm) error {
		if newRatio.lessThan(trustLevelMinAllowed) || trustLevelMaxAllowed.lessThan(newRatio) {
			return errors.New(errInvalidTrustLevel)
		}
		trustLevelRatio = newRatio
		return nil
	}
	return dao.NewSimpleExecutor(callback, "")
}

// GetCooldown returns the currently configured valset-update cooldown,
// expressed in seconds. The default is 24h (per the #4829 acceptance
// criteria); governance can tune it via NewCooldownPropRequest.
func GetCooldown() uint64 {
	return uint64(valsetUpdateCooldown / time.Second)
}

// NewCooldownPropRequest builds a GovDAO proposal that, when executed,
// updates valsetUpdateCooldown. The new duration is given in seconds;
// 0 disables the cooldown entirely (useful for testnets and
// integration tests). The executor reads the cooldown live (not via
// snapshot) so a successful proposal takes effect immediately for any
// valset proposal created after it executes. In-flight valset
// proposals that passed their own create-time cooldown check are
// re-checked at execute time against the (now possibly different)
// live cooldown — this is intentional so governance can both shorten
// AND lengthen the cooldown without surprises.
func NewCooldownPropRequest(seconds uint64, title, description string) dao.ProposalRequest {
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}
	if seconds > cooldownMaxSeconds {
		panic(errInvalidCooldown)
	}
	newCooldown := time.Duration(seconds) * time.Second
	return dao.NewProposalRequest(title, description, newCooldownExecutor(newCooldown))
}

// newCooldownExecutor builds the GovDAO executor that applies a
// cooldown update. Split out for direct testability (ProposalRequest
// doesn't expose its captured executor).
func newCooldownExecutor(newCooldown time.Duration) dao.Executor {
	callback := func(cur realm) error {
		valsetUpdateCooldown = newCooldown
		return nil
	}
	return dao.NewSimpleExecutor(callback, "")
}
