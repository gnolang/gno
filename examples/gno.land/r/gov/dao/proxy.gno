package dao

import (
	"chain"
	"chain/runtime"
	"errors"
	"strconv"

	"gno.land/p/nt/ufmt"
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

// Remember this realm for rendering.
var gRealm = runtime.CurrentRealm()

// Render calls directly to Render's DAO implementation.
// This allows to have this realm as the main entry point for everything.
func Render(p string) string {
	if dao == nil {
		return "DAO not initialized"
	}
	return dao.Render(gRealm.PkgPath(), p)
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
	if dao == nil {
		return false
	}
	execute, err := dao.PreExecuteProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	if !execute {
		return false
	}
	prop, err := GetProposal(cur, pid)
	if err != nil {
		panic(err.Error())
	}
	if err := prop.executor.Execute(cross); err != nil {
		panic(err.Error())
	}
	return true
}

// CreateProposal will try to create a new proposal, that will be validated by the actual
// govDAO implementation. If the proposal cannot be created, an error will be returned.
func CreateProposal(cur realm, r ProposalRequest) (ProposalID, error) {
	if dao == nil {
		return -1, errors.New("DAO not initialized")
	}
	author, err := dao.PreCreateProposal(r)
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
	dao.PostCreateProposal(r, pid)

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
	return dao.VoteOnProposal(r)
}

// MustVoteOnProposalSimple is like MustVoteOnProposal but intended to be used through gnokey with basic types.
func MustVoteOnProposalSimple(cur realm, pid int64, option string) {
	MustVoteOnProposal(cur, VoteRequest{
		Option:     VoteOption(option),
		ProposalID: ProposalID(pid),
	})
}

func MustGetProposal(cur realm, pid ProposalID) *Proposal {
	p, err := GetProposal(cur, pid)
	if err != nil {
		panic(err.Error())
	}

	return p
}

// GetProposal gets created proposal by its ID
func GetProposal(cur realm, pid ProposalID) (*Proposal, error) {
	if dao == nil {
		return nil, errors.New("DAO not initialized")
	}
	if err := dao.PreGetProposal(pid); err != nil {
		return nil, err
	}

	prop := proposals.GetProposal(pid)
	if prop == nil {
		return nil, errors.New(ufmt.Sprintf("Proposal %v does not exist.", int64(pid)))
	}

	if err := dao.PostGetProposal(pid, prop); err != nil {
		return nil, err
	}

	return prop, nil
}

// UpdateImpl is a method intended to be used on a proposal.
// This method will update the current govDAO implementation
// to a new one. AllowedDAOs are a list of realms that can
// call this method, in case the new DAO implementation had
// a breaking bug. Any value set as nil will be ignored.
// If AllowedDAOs field is not set correctly, the actual DAO
// implementation wont be able to execute new Proposals!
func UpdateImpl(cur realm, r UpdateRequest) {
	gRealm := runtime.PreviousRealm().PkgPath()

	if !InAllowedDAOs(gRealm) {
		panic("permission denied for prev realm: " + gRealm)
	}

	if r.AllowedDAOs != nil {
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
