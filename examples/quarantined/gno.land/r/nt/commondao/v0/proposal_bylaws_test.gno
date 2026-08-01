package commondao

import (
	"strings"
	"testing"

	"gno.land/p/nt/bylaws/v0"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	_ commondao.ProposalDefinition = amendBylawsProposal{}
	_ commondao.Validable          = amendBylawsProposal{}
	_ commondao.Executable         = amendBylawsProposal{}
)

// TestAmendBylawsKindNew pins the amend-bylaws factory: args validation,
// the host-identity pin, the stale-at-create fail-fast, the no-op reject,
// and the malformed-script reject. Without this, removing any of those
// guards is invisible to the suite.
func TestAmendBylawsKindNew(t *testing.T) {
	dao := newTestDAO(t, "BylawsHost")
	other := newTestDAO(t, "BylawsOther")
	set := bylawsOf(dao.ID())

	// Wrong args type is a realm bug, reported as invalid args.
	_, err := amendBylawsKind{}.New(dao.Readonly(), textArgs{})
	uassert.ErrorIs(t, err, errInvalidProposalArgs)

	// A nil document set is a realm bug too.
	freshPatch := func(text string) bylaws.Patch {
		t.Helper()
		p, err := set.Diff("bylaws/a.md", text)
		urequire.NoError(t, err)
		return p
	}
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), patch: freshPatch("one")})
	uassert.ErrorIs(t, err, errInvalidProposalArgs)

	// Host-identity pin: a set/daoID for a different DAO than the one being
	// proposed on is rejected.
	_, err = amendBylawsKind{}.New(other.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: freshPatch("one")})
	uassert.ErrorIs(t, err, errInvalidProposalArgs)

	// Stale-at-create: a patch whose base does not match the current
	// document is rejected up front.
	stale := bylaws.Patch{Path: "bylaws/a.md", Base: bylaws.HashText("not the base"), Ops: []bylaws.Op{{Type: bylaws.OpInsert, Text: "x"}}}
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: stale})
	uassert.ErrorIs(t, err, bylaws.ErrStalePatch)

	// No-op amendments are rejected: a council vote is always about an
	// actual change.
	noop, err := bylaws.DiffTexts("bylaws/a.md", "", "", false)
	urequire.NoError(t, err)
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: noop})
	urequire.Error(t, err)
	uassert.Equal(t, "bylaws amendment must change the document", err.Error())

	// A fresh-base patch with a script that does not fit the document is
	// rejected (Format validates the script while rendering the display).
	// The op must not be Keep-only, or the no-op check fires first.
	broken := bylaws.Patch{Path: "bylaws/a.md", Base: "", Ops: []bylaws.Op{{Type: bylaws.OpDelete, N: 5}}}
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: broken})
	uassert.ErrorIs(t, err, bylaws.ErrInvalidPatch)

	// Happy path: the definition renders the change, titles by action, and
	// applies the simple-majority threshold.
	def, err := amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: freshPatch("Quorum is half.")})
	urequire.NoError(t, err)
	abp, ok := def.(amendBylawsProposal)
	urequire.True(t, ok, "expect an amendBylawsProposal")
	uassert.Equal(t, "Add Bylaws Document: bylaws/a.md", abp.Title())
	uassert.True(t, strings.Contains(abp.Body(), "Quorum is half."), "expect body to show the inserted text")
	uassert.True(t, abp.Threshold() == commondao.ThresholdSimpleMajority, "expect simple majority")
	urequire.True(t, abp.Executor() != nil, "expect a non-nil executor")
}

// TestAmendBylawsLifecycle pins the executor and the stale re-check: the
// definition applies its patch on execution, and a competing amendment
// that lost the race fails cleanly without clobbering.
func TestAmendBylawsLifecycle(t *testing.T) {
	dao := newTestDAO(t, "BylawsLifecycle")
	set := bylawsOf(dao.ID())

	newDef := func(text string) amendBylawsProposal {
		t.Helper()
		p, err := set.Diff("mandates/spend.md", text)
		urequire.NoError(t, err)
		def, err := amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: p})
		urequire.NoError(t, err)
		return def.(amendBylawsProposal)
	}

	// Create the document through the executor.
	create := newDef("Spend only by vote.")
	uassert.NoError(t, create.Validate())
	urequire.NoError(t, create.execute(0, nil))
	text, _ := set.Get("mandates/spend.md")
	uassert.Equal(t, "Spend only by vote.", text)

	// Two amendments race from the same base: the first applies, the
	// second is stale at Validate AND at execute, and clobbers nothing.
	winner := newDef("Spend only by supermajority vote.")
	loser := newDef("Spend freely.")
	urequire.NoError(t, winner.execute(0, nil))
	uassert.ErrorIs(t, loser.Validate(), bylaws.ErrStalePatch)
	uassert.ErrorIs(t, loser.execute(0, nil), bylaws.ErrStalePatch)
	text, _ = set.Get("mandates/spend.md")
	uassert.Equal(t, "Spend only by supermajority vote.", text)

	// Removal: an empty proposed text removes the document.
	remove := newDef("")
	uassert.Equal(t, "Remove Bylaws Document: mandates/spend.md", remove.Title())
	urequire.NoError(t, remove.execute(0, nil))
	uassert.False(t, set.Has("mandates/spend.md"))
}
