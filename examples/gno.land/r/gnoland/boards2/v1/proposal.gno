package boards2

import (
	"strings"

	"gno.land/p/nt/ufmt/v0"
	"gno.land/r/gov/dao"
)

// ProposeCreateBoard returns a GovDAO proposal request that creates a new board.
// Only GovDAO members can submit the request (e.g. dao.MustCreateProposal), so owner
// is explicit rather than taken from the caller.
func ProposeCreateBoard(cur realm, name string, listed, open bool, owner address) dao.ProposalRequest {
	if !cur.IsCurrent() {
		panic("ProposeCreateBoard must be called as a crossing function")
	}

	// Fail before a doomed proposal reaches a vote. createBoard re-checks
	// whatever can still change before the proposal is executed.
	assertRealmIsNotLocked()
	assertMemberAddressIsValid(owner)

	name = strings.TrimSpace(name)
	assertIsValidBoardName(name)
	assertBoardNameNotExists(name)

	// Same pair validateBasicBoardCreate applies; bypassing gPerms bypasses it.
	if err := checkBoardNameIsNotAddress(name); err != nil {
		panic(err.Error())
	}
	if err := checkBoardNameBelongsToAddress(owner, name); err != nil {
		panic(err.Error())
	}

	// Returns an error rather than panicking which would abort the
	// execution transaction and leave the proposal active.
	cb := func(realm) error {
		// Note: This advances the global ID counter, even on error
		return createBoard(gBoardsSequence.Next(), owner, name, listed, open)
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")

	// Board names are constrained to reBoardName, so they carry no markdown.
	return dao.NewProposalRequest(
		ufmt.Sprintf(`Boards2 V1: Create board "%s"`, name),
		ufmt.Sprintf(
			"Create a new board in [%s](%s).\n\n- Owner: %s\n- Listed: %t\n- Open: %t",
			realmPkgPath, gRealmPath, owner, listed, open,
		),
		e,
	)
}
