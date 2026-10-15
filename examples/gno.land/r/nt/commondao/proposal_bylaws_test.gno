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
	_ trustedMarkdownBody          = amendBylawsProposal{}
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

	// Set-identity pin: the args-captured set must be the host's canonical
	// set — a matching daoID with a different DAO's set is rejected, so a
	// wrapper can never validate one DAO's documents and amend another's.
	otherSet := bylawsOf(other.ID())
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: otherSet, patch: freshPatch("one")})
	uassert.ErrorIs(t, err, errInvalidProposalArgs)

	// Invalid document paths fail fast at creation.
	badPath := bylaws.Patch{Path: "bad path", Base: "", Ops: []bylaws.Op{{Type: bylaws.OpInsert, Text: "x"}}}
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: badPath})
	uassert.ErrorIs(t, err, bylaws.ErrInvalidPath)

	// The mandates/ folder is reserved: mandates are set at creation or by
	// an ancestor, never by the council's own amendment power.
	mandate, err := bylaws.DiffTexts("mandates/spend.md", "", "Spend only by vote.", false)
	urequire.NoError(t, err)
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: mandate})
	urequire.Error(t, err)
	uassert.True(t, strings.Contains(err.Error(), "mandates are not council-amendable"), "expect the mandates reservation")
	bare, err := bylaws.DiffTexts("mandates", "", "x", false)
	urequire.NoError(t, err)
	_, err = amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: bare})
	urequire.Error(t, err)
	uassert.True(t, strings.Contains(err.Error(), "mandates are not council-amendable"), "expect the bare-path reservation")

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
	// applies the supermajority default threshold.
	def, err := amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: freshPatch("Quorum is half.")})
	urequire.NoError(t, err)
	abp, ok := def.(amendBylawsProposal)
	urequire.True(t, ok, "expect an amendBylawsProposal")
	uassert.Equal(t, "Add Bylaws Document: bylaws/a.md", abp.Title())
	uassert.True(t, strings.Contains(abp.Body(), "Quorum is half."), "expect body to show the inserted text")
	uassert.True(t, strings.Contains(abp.Body(), "bylaws/a\\.md"), "expect the path escaped in the body")
	uassert.True(t, abp.Threshold() == commondao.ThresholdSupermajority, "expect the supermajority default")
	urequire.True(t, abp.Executor() != nil, "expect a non-nil executor")
}

// TestAmendBylawsIsolation pins that amendments are keyed per DAO: two
// DAOs hold the same path independently and amending one never touches
// the other.
func TestAmendBylawsIsolation(t *testing.T) {
	dao1 := newTestDAO(t, "IsoOne")
	dao2 := newTestDAO(t, "IsoTwo")
	set1, set2 := bylawsOf(dao1.ID()), bylawsOf(dao2.ID())

	p1, err := set1.Diff("bylaws/x.md", "text one")
	urequire.NoError(t, err)
	urequire.NoError(t, set1.Apply(p1))

	p2, err := set2.Diff("bylaws/x.md", "text two")
	urequire.NoError(t, err)
	urequire.NoError(t, set2.Apply(p2))

	t1, _ := set1.Get("bylaws/x.md")
	t2, _ := set2.Get("bylaws/x.md")
	uassert.Equal(t, "text one", t1)
	uassert.Equal(t, "text two", t2)

	// Amend one; the other is untouched.
	p1b, err := set1.Diff("bylaws/x.md", "")
	urequire.NoError(t, err)
	urequire.NoError(t, set1.Apply(p1b))
	uassert.False(t, set1.Has("bylaws/x.md"))
	t2, _ = set2.Get("bylaws/x.md")
	uassert.Equal(t, "text two", t2)
}

// TestAmendBylawsLifecycle pins the executor and the stale re-check: the
// definition applies its patch on execution, and a competing amendment
// that lost the race fails cleanly without clobbering.
func TestAmendBylawsLifecycle(t *testing.T) {
	dao := newTestDAO(t, "BylawsLifecycle")
	set := bylawsOf(dao.ID())

	newDef := func(text string) amendBylawsProposal {
		t.Helper()
		p, err := set.Diff("bylaws/spend.md", text)
		urequire.NoError(t, err)
		def, err := amendBylawsKind{}.New(dao.Readonly(), amendBylawsProposal{daoID: dao.ID(), set: set, patch: p})
		urequire.NoError(t, err)
		return def.(amendBylawsProposal)
	}

	// Create the document through the executor.
	create := newDef("Spend only by vote.")
	uassert.NoError(t, create.Validate())
	urequire.NoError(t, create.execute(0, nil))
	text, _ := set.Get("bylaws/spend.md")
	uassert.Equal(t, "Spend only by vote.", text)

	// Two amendments race from the same base: the first applies, the
	// second is stale at Validate AND at execute, and clobbers nothing.
	winner := newDef("Spend only by supermajority vote.")
	loser := newDef("Spend freely.")
	urequire.NoError(t, winner.execute(0, nil))
	uassert.ErrorIs(t, loser.Validate(), bylaws.ErrStalePatch)
	uassert.ErrorIs(t, loser.execute(0, nil), bylaws.ErrStalePatch)
	text, _ = set.Get("bylaws/spend.md")
	uassert.Equal(t, "Spend only by supermajority vote.", text)

	// Removal: an empty proposed text removes the document.
	remove := newDef("")
	uassert.Equal(t, "Remove Bylaws Document: bylaws/spend.md", remove.Title())
	urequire.NoError(t, remove.execute(0, nil))
	uassert.False(t, set.Has("bylaws/spend.md"))
}
