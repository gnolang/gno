package dao

import (
	"chain"
	"errors"
	"strconv"
	"strings"

	"gno.land/p/nt/ufmt/v0"
)

// dao is the actual govDAO implementation, having all the needed business logic
var dao DAO

// allowedDAOs contains realms that can be used to update the actual govDAO implementation,
// and validate Proposals.
// This is like that to be able to rollback using a previous govDAO implementation in case
// the latest implementation has a breaking bug. After a test period, a proposal can be
// executed to remove all previous govDAOs implementations and leave the last one.
var allowedDAOs []string

// proposals contains all the proposals in history.
var proposals *Proposals = NewProposals()

// Render calls directly to Render's DAO implementation.
// This allows to have this realm as the main entry point for everything.
func Render(cur realm, p string) string {
	if dao == nil {
		return "DAO not initialized"
	}
	return dao.Render(cross(cur), cur.PkgPath(), p)
}

// MustCreateProposal is an utility method that does the same as CreateProposal,
// but instead of erroing if something happens, it panics.
func MustCreateProposal(cur realm, r ProposalRequest) ProposalID {
	pid, err := CreateProposal(cur, r)
	if err != nil {
		panic(err.Error())
	}

	return pid
}

// ExecuteProposal will try to execute the proposal with the provided ProposalID.
// If the proposal was denied, it will return false. If the proposal is correctly
// executed, it will return true. If something happens this function will panic.
func ExecuteProposal(cur realm, pid ProposalID) bool {
	return executeProposal(cur, pid, false)
}

// ExecuteOrRejectProposal executes the proposal with the provided ProposalID or rejects
// it when there is an execution error.
// If the proposal was denied, it will return false. If the proposal is correctly
// executed, it will return true, unless execution fails with an error, in which case
// proposal is rejected with the error as the reason.
// This function allows to finish proposals by rejecting them when there is a state
// change or an error in the proposal parameters that makes execution fail, potentially
// leaving the proposal active forever because it can't be successfully executed.
func ExecuteOrRejectProposal(cur realm, pid ProposalID) bool {
	return executeProposal(cur, pid, true)
}

// CreateProposal will try to create a new proposal, that will be validated by the actual
// govDAO implementation. If the proposal cannot be created, an error will be returned.
func CreateProposal(cur realm, r ProposalRequest) (ProposalID, error) {
	if dao == nil {
		return -1, errors.New("DAO not initialized")
	}
	author, err := dao.PreCreateProposal(0, cur, r)
	if err != nil {
		return -1, err
	}

	p := &Proposal{
		author:      author,
		title:       r.title,
		description: r.description,
		executor:    r.executor,
		allowedDAOs: allowedDAOs[:],
	}

	pid := proposals.SetProposal(p)
	dao.PostCreateProposal(0, cur, r, pid)

	chain.Emit("ProposalCreated",
		"id", strconv.FormatInt(int64(pid), 10),
	)

	return pid, nil
}

func MustVoteOnProposal(cur realm, r VoteRequest) {
	if err := VoteOnProposal(cur, r); err != nil {
		panic(err.Error())
	}
}

// VoteOnProposal sends a vote to the actual govDAO implementation.
// If the voter cannot vote the specified proposal, this method will return an error
// with the explanation of why.
func VoteOnProposal(cur realm, r VoteRequest) error {
	if dao == nil {
		return errors.New("DAO not initialized")
	}
	return dao.VoteOnProposal(0, cur, r)
}

// MustVoteOnProposalSimple is like MustVoteOnProposal but intended to be used through gnokey with basic types.
func MustVoteOnProposalSimple(cur realm, pid int64, option string) {
	MustVoteOnProposal(cur, VoteRequest{
		Option:     VoteOption(option),
		ProposalID: ProposalID(pid),
	})
}

func MustGetProposal(pid ProposalID) *Proposal {
	p, err := GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	return p
}

// GetProposal gets created proposal by its ID. Non-crossing pure read:
// looks up the proposal in this realm's package var. Callable directly
// from any realm without cross-call syntax.
func GetProposal(pid ProposalID) (*Proposal, error) {
	if dao == nil {
		return nil, errors.New("DAO not initialized")
	}
	prop := proposals.GetProposal(pid)
	if prop == nil {
		return nil, errors.New(ufmt.Sprintf("Proposal %v does not exist.", int64(pid)))
	}
	return prop, nil
}

// UpdateImpl is a method intended to be used on a proposal.
// This method will update the current govDAO implementation
// to a new one. AllowedDAOs are a list of realms that can
// call this method, in case the new DAO implementation had
// a breaking bug. A nil DAO is ignored.
// If AllowedDAOs field is not set correctly, the actual DAO
// implementation wont be able to execute new Proposals!
//
// An empty AllowedDAOs is ignored rather than stored. An empty list makes
// InAllowedDAOs() return true for every caller — the bootstrap-only state that
// lets the genesis MsgRun seed the member set. Since this is the only site that
// assigns allowedDAOs, ignoring empty here makes the transition
// empty -> non-empty one-way: once locked down the DAO cannot be reopened,
// whether the empty value arrives as a literal []string{} or from
// NewUpdateRequest(d, nil), which copies nil into a non-nil empty slice.
// Individual entries must be non-blank realm paths; an empty entry would match
// a user realm's empty PkgPath() and is rejected.
func UpdateImpl(cur realm, r UpdateRequest) {
	gRealm := cur.Previous().PkgPath()

	if !InAllowedDAOs(gRealm) {
		panic("permission denied for prev realm: " + gRealm)
	}

	if len(r.AllowedDAOs) != 0 {
		// Every entry must be a real realm path. len() != 0 alone is the wrong
		// invariant: InAllowedDAOs compares by exact string, and a user realm's
		// PkgPath() is "", so a single "" entry admits any caller whose previous
		// frame is a user realm -- the same fail-open outcome this guard exists
		// to prevent, just spelled differently. An empty entry can only be a
		// drafting mistake, so reject the whole request rather than silently
		// dropping it and storing a list the proposal did not describe.
		for i, d := range r.AllowedDAOs {
			trimmed := strings.TrimSpace(d)
			if trimmed == "" {
				panic("AllowedDAOs entries must be realm paths; got an empty one")
			}
			// Entries are stored exactly as given, and InAllowedDAOs compares
			// whole strings, so an entry with surrounding spaces matches no
			// caller at all. A non-empty list also closes the bootstrap
			// window, so a list of only padded entries is locked shut against
			// everyone, the DAO included, with no way to reopen it.
			//
			// This does not make the list typo-proof, and is not meant to be:
			// any wrong path locks the DAO out exactly the same way, and no
			// check here can tell a typo from a realm that does not exist yet.
			// Whitespace is worth rejecting because it is the one spelling a
			// human reviewing the proposal cannot see. Rejected rather than
			// trimmed, so what gets stored is what the proposal said.
			// Reported by position, not by value. A panic message becomes the
			// proposal's DeniedReason, which is stored, and the entry is
			// caller-supplied and unbounded — echoing it would put an
			// arbitrary amount of someone else's text into this realm's
			// storage. The index is enough to find it in a list the proposal
			// author wrote.
			if d != trimmed {
				panic("AllowedDAOs entries must not have leading or trailing spaces; entry " + strconv.Itoa(i))
			}
		}
		// Stored as given. A defensive copy here looks prudent but would
		// guard a write no outside realm can perform, which was checked from
		// a separate realm rather than assumed:
		//
		//   - Building this request as a literal fails outright, with
		//     "cannot allocate gno.land/r/gov/dao.UpdateRequest in realm ...".
		//   - Writing through a request obtained from NewUpdateRequest fails
		//     with "cannot directly modify readonly tainted object".
		//
		// So the only way in is NewUpdateRequest, which copies already. Both
		// checks are language guarantees, not conventions.
		allowedDAOs = r.AllowedDAOs
	}

	if r.DAO != nil {
		dao = r.DAO
	}
}

func AllowedDAOs() []string {
	dup := make([]string, len(allowedDAOs))
	copy(dup, allowedDAOs)
	return dup
}

func InAllowedDAOs(pkg string) bool {
	if len(allowedDAOs) == 0 {
		return true // corner case for initialization
	}
	for _, d := range allowedDAOs {
		if pkg == d {
			return true
		}
	}
	return false
}

func executeProposal(cur realm, pid ProposalID, execErrorRejects bool) bool {
	if dao == nil {
		return false
	}
	execute, err := dao.PreExecuteProposal(0, cur, pid)
	if err != nil {
		panic(err.Error())
	}

	if !execute {
		return false
	}
	prop, err := GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	err = dao.ExecuteProposal(0, cur, pid, prop.executor)
	if err != nil {
		if execErrorRejects {
			return false
		}

		panic(err.Error())
	}
	return true
}
