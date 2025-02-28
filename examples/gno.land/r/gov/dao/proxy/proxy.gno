package proxy

import "std"

// dao is the actual govDAO implementation, having all the needed business logic
var dao DAO

// AllowedDAOs contains realms that can be used to update the actual govDAO implementation,
// and validate Proposals.
// This is like that to be able to rollback using a previous govDAO implementation in case
// the latest implementation has a breaking bug. After a test period, a proposal can be
// executed to remove all previous govDAOs implementations and leave the last one.
var AllowedDAOs []string

// proposals contains all the proposals in history.
var proposals []*Proposal

func init() {
	AllowedDAOs = []string{"gno.land/r/gov/dao/v3/impl"}
}

// Render calls directly to Render's DAO implementation.
// This allows to have this realm as the main entry point for everything.
func Render(p string) string {
	return dao.Render(p)
}

// MustCreateProposal is an utility method that does the same as CreateProposal,
// but instead of erroing if something happens, it panics.
func MustCreateProposal(r ProposalRequest) ProposalID {
	pid, err := CreateProposal(r)
	if err != nil {
		panic(err.Error())
	}

	return pid
}

// CreateProposal will try to create a new proposal, that will be validated by the actual
// govDAO implementation. If the proposal cannot be created, an error will be returned.
func CreateProposal(r ProposalRequest) (ProposalID, error) {
	//TODO: get prev realm here and remove the author coming from the govDAO impl
	author, err := dao.PreCreateProposal(r)
	if err != nil {
		return -1, err
	}

	p := &Proposal{
		Author:      author,
		Title:       r.Title,
		Description: r.Description,
		Executor:    newExecutor(r.Callback, r.Metadata),
		AllowedDAOs: AllowedDAOs,
	}

	pid := ProposalID(len(proposals))

	proposals = append(proposals, p)

	dao.PostCreateProposal(r, pid)

	return pid, nil
}

func MustVoteOnProposal(r VoteRequest) {
	if err := VoteOnProposal(r); err != nil {
		panic(err.Error())
	}
}

// VoteOnProposal sends a vote to the actual govDAO implementation.
// If the voter cannot vote the specified proposal, this method will return an error
// with the explanation of why.
func VoteOnProposal(r VoteRequest) error {
	return dao.VoteOnProposal(r)
}

func MustGetProposal(pid ProposalID) *Proposal {
	p, err := GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	return p
}

// GetProposal gets created proposal by its ID
func GetProposal(pid ProposalID) (*Proposal, error) {
	if err := dao.PreGetProposal(pid); err != nil {
		return nil, err
	}

	ipid := int(pid)
	if len(proposals) <= ipid {
		return nil, nil
	}

	prop := proposals[ipid]

	if err := dao.PostGetProposal(pid, prop); err != nil {
		return nil, err
	}

	return proposals[ipid], nil
}

// UpdateImpl is a method intended to be used on a proposal.
// This method will update the current govDAO implementation
// to a new one. newDAOUpdaters are a list of realms that can
// call this method, in case the new DAO implementation had
// a breaking bug. Any value set as nil will be ignored.
// If AllowedDAOs field is not set correctly, the actual DAO
// implementation wont be able to execute new Proposals!
func UpdateImpl(r UpdateRequest) {
	realm := std.PreviousRealm().PkgPath()
	if index(realm) < 0 {
		panic("permission denied for prev realm: " + realm)
	}

	if r.AllowedDAOs != nil {
		AllowedDAOs = r.AllowedDAOs
	}

	if r.DAO != nil {
		dao = r.DAO
	}
}

func InAllowedDAOsIndex(pkg string) int {
	return index(pkg)
}

func index(pkg string) int {
	for i, d := range AllowedDAOs {
		if pkg == d {
			return i
		}
	}

	return -1
}
