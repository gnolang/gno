package commondao

import (
	"errors"
	"time"

	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/bylaws/v0"
	"gno.land/p/nt/commondao/v0"
)

// bylawsSets stores each DAO's governing documents (bylaws and
// mandates), keyed like daos. A set is created lazily on a DAO's first
// amendment proposal; a DAO with no documents has no entry. The
// *bylaws.Bylaws handle is mutable (Apply amends it), so it never leaves
// the realm — public reads return strings.
var bylawsSets = bptree.NewBPTree32() // string(ID) -> *bylaws.Bylaws

// bylawsView returns a DAO's document set or nil when it has none. Read
// paths (render, public getters, payload building) use this — it never
// writes realm state, so it is safe under read-only query evaluation.
func bylawsView(daoID uint64) *bylaws.Bylaws {
	if v := bylawsSets.Get(makeIDKey(daoID)); v != nil {
		return v.(*bylaws.Bylaws)
	}
	return nil
}

// bylawsOf returns a DAO's document set, creating an empty one on first
// use. Only proposal creation calls it; an empty set left behind by a
// failed proposal is harmless.
func bylawsOf(daoID uint64) *bylaws.Bylaws {
	if b := bylawsView(daoID); b != nil {
		return b
	}
	b := bylaws.New()
	bylawsSets.Set(makeIDKey(daoID), b)
	return b
}

// amendBylawsProposal is both the amend-bylaws args struct and the
// proposal definition it produces (one type serves both roles, like
// manageKindsProposal). set is the host DAO's document set the executor
// patches: New receives only a readonly view, so the trusted wrapper
// passes the mutable set through args (from its own bylawsOf) together
// with the host daoID for the identity pin. display is the human-readable
// change, rendered at New against the then-current document.
type amendBylawsProposal struct {
	daoID   uint64
	set     *bylaws.Bylaws
	patch   bylaws.Patch
	display string
}

// amendBylawsKind creates proposals that add, amend or remove one of the
// host DAO's bylaws/mandates documents with a verifiable diff patch. The
// patch pins the sha256 of the document text it was diffed against, so a
// passed amendment that raced a concurrent change to the same document
// fails cleanly (StatusFailed) instead of clobbering it. Decided by
// simple majority: the Constitution sets that bar for charter, bylaws
// and mandate changes.
type amendBylawsKind struct{}

func (amendBylawsKind) Name() string { return kindAmendBylaws }

func (amendBylawsKind) New(dao commondao.ReadonlyCommonDAO, args any) (commondao.ProposalDefinition, error) {
	p, ok := args.(amendBylawsProposal)
	if !ok || p.set == nil {
		return nil, errInvalidProposalArgs
	}
	// Defense in depth: pin the args-captured document set to the readonly
	// host Propose passed here, so a future wrapper can never validate
	// against one DAO's documents and amend another's (the trusted wrapper
	// always passes matching handles today).
	if p.daoID != dao.ID() {
		return nil, errInvalidProposalArgs
	}

	// Freshness fail-fast: reject a patch that is already stale at
	// creation (Validate re-checks at execution).
	cur, _ := p.set.Get(p.patch.Path)
	if p.set.Hash(p.patch.Path) != p.patch.Base {
		return nil, bylaws.ErrStalePatch
	}
	if p.patch.IsNoop() {
		return nil, errors.New("bylaws amendment must change the document")
	}

	// Rendering the change also validates the edit script against the
	// current text, so a malformed patch never becomes a proposal.
	display, err := p.patch.Format(cur)
	if err != nil {
		return nil, err
	}
	p.display = display
	return p, nil
}

// Title returns the proposal title as raw text: the renderer escapes
// every definition title.
func (p amendBylawsProposal) Title() string {
	verb := "Amend"
	switch {
	case p.patch.IsCreate():
		verb = "Add"
	case p.patch.IsRemove():
		verb = "Remove"
	}
	return verb + " Bylaws Document: " + p.patch.Path
}

// Body returns the change summary as raw text (document content and
// proposer-supplied insertions), so the renderer escapes it like the
// execution kind's body.
func (p amendBylawsProposal) Body() string {
	return "Document: " + p.patch.Path + "\n\n" + p.display
}

func (amendBylawsProposal) VotingPeriod() time.Duration { return time.Hour * 24 * 7 }

// Threshold returns the tally threshold: the Constitution decides
// charter, bylaws and mandate changes by simple majority.
func (amendBylawsProposal) Threshold() commondao.Threshold {
	return commondao.ThresholdSimpleMajority
}

// Validate re-asserts patch freshness at execution (Validate reruns
// inside Execute): a document amended after this proposal passed fails
// it cleanly (StatusFailed) instead of clobbering the newer text.
func (p amendBylawsProposal) Validate() error {
	if p.set.Hash(p.patch.Path) != p.patch.Base {
		return bylaws.ErrStalePatch
	}
	return nil
}

func (p amendBylawsProposal) Executor() commondao.ExecFunc {
	return p.execute
}

// execute applies the patch, returning any bylaws error unchanged so a
// race between two passed amendments fails the later one cleanly
// (StatusFailed) instead of panicking the transaction. It moves no
// funds, so the definition is not Funded and ignores sub.
func (p amendBylawsProposal) execute(_ int, _ realm) error {
	return p.set.Apply(p.patch)
}
