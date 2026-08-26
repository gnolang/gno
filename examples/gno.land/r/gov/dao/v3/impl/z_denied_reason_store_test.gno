package impl

import (
	"errors"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/memberstore"
)

// Named z_* so it sorts after govdao_test.gno, whose TestUpgradeDaoImplementation
// asserts a hard-coded proposal id. This builds its own GovDAO instance and
// registers a status directly, so it never touches the shared one.

type hugeErrExecutor struct{ msg string }

func (e *hugeErrExecutor) Execute(cur realm) error { return errors.New(e.msg) }

func (e *hugeErrExecutor) String() string { return "" }

func (e *hugeErrExecutor) CreationRealm() string { return "gno.land/r/test/hugeerr" }

// The denial reason is stored, not just shown. It is "execution failed: " plus
// whatever error the proposal's executor returned, and the executor is
// third-party code. Whoever executes the proposal pays the storage deposit for
// it, and that need not be the person who wrote the executor.
//
// This has to read the stored value rather than the rendered page. The render
// clamps to the same bound, so a page rendered from an unclamped store looks
// exactly like one rendered from a clamped store — testing through Render would
// pass whether or not the write is bounded.
func TestDeniedReasonIsClampedBeforeStoring(cur realm, t *testing.T) {
	g := NewGovDAO()
	pid := dao.ProposalID(987654)
	g.pss.Set(pid.String(), newProposalStatus([]string{memberstore.T1}))

	g.ExecuteProposal(0, cur, pid, &hugeErrExecutor{msg: strings.Repeat("E", 50000)})

	stored := g.pss.GetStatus(pid).DeniedReason
	uassert.True(t, len(stored) < maxRenderedReason+64,
		"the realm must not store a reason larger than it can display")
	uassert.True(t, strings.Contains(stored, "… truncated"),
		"a stored reason that was cut must say so")
	uassert.True(t, strings.HasPrefix(stored, "execution failed: E"),
		"the prefix and the start of the error must survive")
}
