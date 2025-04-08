package dao

import (
	"errors"
	"math"
	"std"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/pol/dao"
	gm "gno.land/r/pol/gnominers"
)

const (
	DaoNo dao.ProposalChoice = iota
	DaoYes
	_DaoChoiceCount
)

type MinerExecutor struct {
	session *gm.UserInfo // The session
	count   uint         // How many gems you use in from
	from    gm.GemType   // The gem you want to give
	to      gm.GemType   // The gem you want to convert
}

func safelyRemoveGems(u *gm.UserInfo, gem gm.GemType, count uint) {
	u.GemsCount[gem] = uint(math.Max(0, float64(u.GemsCount[gem])-float64(count)))
}

func merge(diff float64, fromCount uint) float64 {
	ratio := float64(fromCount) / math.Pow(10, diff)
	prob := ratio / (ratio + 1)
	if prob > 1 {
		return 1
	}
	return prob
}

func unMerge(diff float64, fromCount uint) uint {
	return uint(math.Pow(10, diff)) * fromCount
}

func (e *MinerExecutor) Execute() error {
	if e.session.GemsCount[e.from] < e.count {
		return errors.New("not enough gems")
	}
	switch {
	case e.from == e.to:
		return errors.New("cant convert gems of the same type")
	case e.from < e.to:
		prob := merge(float64(e.to-e.from), e.count)
		safelyRemoveGems(e.session, e.from, e.count)
		if gm.RandomMachine.RandFloat() <= prob {
			e.session.GemsCount[e.to] += 1
		}
	case e.from > e.to:
		safelyRemoveGems(e.session, e.from, e.count)
		e.session.GemsCount[e.to] += uint(math.Round(float64(unMerge(float64(e.from-e.to), e.count))))
	}
	return nil
}

type MinerProposal struct {
	session  *gm.UserInfo   // The session
	author   dao.Member     // The original author.
	votes    [2]uint        // The actual votes.
	voters   []dao.Member   // The ones who alread votes.
	executor *MinerExecutor // Execute the poll
	alive    bool           // If the poll is alive
}

func (p *MinerProposal) GetAuthor() dao.Member {
	return p.author
}

func (p *MinerProposal) HowManyVoters() uint {
	return p.votes[DaoYes] + p.votes[DaoNo]
}

func (p *MinerProposal) IsAccepted() bool {
	return p.votes[DaoYes] > p.votes[DaoNo]
}

func (p *MinerProposal) ExecuteIfAccepted() error {
	if p.IsAccepted() {
		return p.executor.Execute()
	}
	return nil
}

func (p *MinerProposal) ExpiresIn() time.Time {
	return time.Unix(0, 0)
}

func (p *MinerProposal) Vote(verdict dao.ProposalChoice) error {
	who := std.PreviousRealm().Address()
	if who.String() == "" || verdict >= _DaoChoiceCount {
		return errors.New("invalid previous realm")
	}
	for _, v := range p.voters {
		if v.Address() == who {
			return errors.New("user already voted")
		}
	}
	p.voters = append(p.voters, dao.Member(who))
	p.votes[verdict] += 1
	return nil
}

func (p MinerProposal) IsEnded() bool {
	return !p.alive
}

func (d MinerProposal) GetChoicesNames() map[dao.ProposalChoice]string {
	return map[dao.ProposalChoice]string{
		DaoYes: "Yes",
		DaoNo:  "No",
	}
}

func (d MinerProposal) String() string {
	var out, s string
	out += "### " + d.author.Address().String() + "'s proposal\n"
	if d.executor.count > 1 {
		s = "s"
	}
	out += ufmt.Sprintf("Convert %d %s%s into %ss",
		d.executor.count,
		gm.GemNames[d.executor.from],
		s,
		gm.GemNames[d.executor.to],
	)
	return out
}

// Create a new miner proposal
func NewMinerProposal(session *gm.UserInfo, who dao.Member, from, to gm.GemType, count uint) *MinerProposal {
	x := &MinerProposal{
		author:   who,
		votes:    [2]uint{0, 0},
		voters:   []dao.Member{},
		executor: &MinerExecutor{
			session: session,
			from: from,
			to: to,
			count: count,
		},
	}
	return x
}

// The miner dao
type MinerDAO struct {
	dao.DAO
	props []dao.Proposal
}

// Get the proposals, also update the dao by removing ended proposals
func (d *MinerDAO) GetProposals() []dao.Proposal {
	newProps := []dao.Proposal{}
	for _, p := range d.props {
		if !p.IsEnded() {
			newProps = append(newProps, p)
		}
	}
	d.props = newProps
	return d.props
}

// Propose a new Proposal
func (d *MinerDAO) Propose(prop dao.Proposal) {
	d.props = append(d.props, prop)
}

func NewMinerDAO() *MinerDAO {
	return &MinerDAO{
		dao.NewDAO(),
		[]dao.Proposal{},
	}
}

