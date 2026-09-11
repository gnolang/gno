package boards2

import (
	"errors"
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/gnoland/boards/exts/permissions"
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

// ProposeSetPermissions returns a GovDAO proposal request that sets the
// permissions p for the realm or for a single board. Use the concrete
// permissions.Permissions struct for p (not the boards.Permissions interface)
// so that we can Clone a fresh copy. p must not have validators, which can't be copied.
// Instead, the existing validators in the permissions for boardID are restored at
// proposal execution time. This means that the existing permissions must be the
// default concrete permissions.Permissions .
func ProposeSetPermissions(cur realm, boardID boards.ID, p *permissions.Permissions) dao.ProposalRequest {
	if !cur.IsCurrent() {
		panic("ProposeSetPermissions must be called as a crossing function")
	}

	// Fail before a doomed proposal reaches a vote. setPermissions re-checks
	// whatever can still change before the proposal is executed.
	assertRealmIsNotLocked()
	assertBoardExists(boardID)
	if _, err := getCopyablePermissions(boardID); err != nil {
		panic(err.Error())
	}

	// Make a fresh clone which is independent of the caller's copy.
	owned := permissions.Clone(p)

	// The executor runs under GovDAO authority, so the recorded caller is the
	// member that submitted the request.
	proposer := cur.Previous().Address()

	// Returns an error rather than panicking which would abort the
	// execution transaction and leave the proposal active.
	cb := func(realm) error {
		src, err := getCopyablePermissions(boardID)
		if err != nil {
			return err
		}
		owned.CopyValidators(src)
		return setPermissions(boardID, proposer, owned)
	}

	e := dao.NewSimpleExecutor(0, cur, cb, "")

	if boardID == 0 {
		return dao.NewProposalRequest(
			"Boards2 V1: Set realm permissions",
			ufmt.Sprintf(
				"Replace the realm permissions of [%s](%s), including its members and roles.\n\n- Proposer: %s",
				realmPkgPath, gRealmPath, proposer,
			),
			e,
		)
	}

	// Board names are constrained to reBoardName, so they carry no markdown.
	name := mustGetBoard(boardID).Name
	return dao.NewProposalRequest(
		ufmt.Sprintf(`Boards2 V1: Set permissions of board "%s"`, name),
		ufmt.Sprintf(
			"Replace the permissions of board `%s` in [%s](%s), including its members and roles.\n\n- Board ID: %s\n- Proposer: %s",
			name, realmPkgPath, gRealmPath, boardID, proposer,
		),
		e,
	)
}

// getCopyablePermissions returns the current realm permissions, or the board
// permissions when boardID is not zero, if their validators can be copied.
func getCopyablePermissions(boardID boards.ID) (*permissions.Permissions, error) {
	current := gPerms
	if boardID != 0 {
		board, found := gBoards.Get(boardID)
		if !found {
			return nil, errBoardNotFound
		}
		current = board.Permissions
	}

	src, ok := current.(*permissions.Permissions)
	if !ok {
		return nil, errors.New("current permissions don't support copying validators")
	}
	return src, nil
}
