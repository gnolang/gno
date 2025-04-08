package gnominers

import (
	"errors"
	"math"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/txlink"
	"gno.land/p/pol/dao"
)

const (
	DaoNo dao.ProposalChoice = iota
	DaoYes
	_DaoChoiceCount
)

type minerExecutor struct {
	session *UserSession // The session
	count   uint      // How many gems you use in from
	from    gemType   // The gem you want to give
	to      gemType   // The gem you want to convert
	success uint  	  // If successful
}

func safelyRemoveGems(u *UserSession, gem gemType, count uint) {
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

var glob string

func (e *minerExecutor) Execute() error {
	if e.session.GemsCount[e.from] < e.count {
		return errors.New("not enough gems")
	}
	switch {
	case e.from == e.to:
		return errors.New("cant convert gems of the same type")
	case e.from < e.to:
		prob := merge(float64(e.to-e.from), e.count)
		safelyRemoveGems(e.session, e.from, e.count)
		rd := randomMachine.RandFloat()
		glob = ufmt.Sprintln(rd, prob)
		if rd <= prob {
			e.session.GemsCount[e.to] += 1
			e.success = 1
		}
	case e.from > e.to:
		safelyRemoveGems(e.session, e.from, e.count)
		e.success = uint(math.Round(randomMachine.RandFloat() * float64(unMerge(float64(e.from-e.to), e.count))))
		e.session.GemsCount[e.to] += e.success
	}
	return nil
}

type minerProposal struct {
	session  *UserSession   // The session
	author   dao.Member     // The original author.
	votes    [2]uint        // The actual votes.
	voters   []dao.Member   // The ones who alread votes.
	executor *minerExecutor // Execute the poll
	ended    bool           // If the poll is ended
	alive 	 bool           // If the poll is alive
	index 	 uint			// the proposal index
}

func (p *minerProposal) GetAuthor() dao.Member {
	return p.author
}

func (p *minerProposal) HowManyVoters() uint {
	return p.votes[DaoYes] + p.votes[DaoNo]
}

func (p *minerProposal) IsAccepted() bool {
	return p.votes[DaoYes] > p.votes[DaoNo]
}

func (p *minerProposal) ExecuteIfAccepted() error {
	p.alive = false
	if !p.IsAccepted() {
		return nil
	}
	err := p.executor.Execute()
	if err != nil {
		return err
	}
	p.alive = true
	p.ended = true
	return nil
}

func (p *minerProposal) ExpiresIn() time.Time {
	return time.Unix(0, 0)
}

func (p *minerProposal) Vote(verdict dao.ProposalChoice) error {
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

func (p minerProposal) IsEnded() bool {
	return !p.alive
}

func (d minerProposal) GetChoicesNames() map[dao.ProposalChoice]string {
	return map[dao.ProposalChoice]string{
		DaoYes: "Yes",
		DaoNo:  "No",
	}
}

func upperFirstLetter(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func DaoResponse(name string, choice dao.ProposalChoice) {
	actor := dao.Member(std.PreviousRealm().Address())
	infos := assertCallerExists()
	infos.DAO.AssertUserIsMember(actor)
	i := mustAtoi(name)

	for _, x := range infos.DAO.GetProposals() {
		prop, ok := x.(*minerProposal)
		if !ok {
			continue
		}
		if !prop.alive || i != int(prop.index) {
			continue
		}
		prop.Vote(choice)
		if prop.HowManyVoters() != uint(len(infos.DAO.GetMembers())) {
			continue
		}
		err := prop.ExecuteIfAccepted()
		if err != nil {
			panic(err)
		}
		break
	}
}

func DaoRespondYes(name string) {
	DaoResponse(name, DaoYes)
}

func DaoRespondNo(name string) {
	DaoResponse(name, DaoNo)
}

func (d minerProposal) String() string {
	if !d.alive {
		return ""
	}
	if d.ended {
		return ufmt.Sprintf("> #%d results with %d %s.", d.index, d.executor.success, upperFirstLetter(gemNames[d.executor.to])) + ", " + glob
	}

	var out, s, quote string
	out += "### " + generateCoolName(d.author.Address()) + "'s proposal n°" + strconv.Itoa(int(d.index)) + "\n"
	if d.executor.count > 1 {
		s = "s"
	}
	if d.executor.from < d.executor.to {
		quote = " (" + strconv.FormatFloat(merge(float64(d.executor.to-d.executor.from), d.executor.count) * 100., 'f', -1, 64) + "% chance)"
	} else {
		quote = " (between 0 and " + strconv.Itoa(int(unMerge(float64(d.executor.from-d.executor.to), d.executor.count))) + ")"
	}
	out += ufmt.Sprintf("Should we convert %d %s%s into %s%s ?\n",
		d.executor.count,
		upperFirstLetter(gemNames[d.executor.from]),
		s,
		upperFirstLetter(gemNames[d.executor.to]),
		quote,
	)
	out += "\n\n"
	out += strconv.Itoa(int(d.HowManyVoters())) + " votes\n\n"
	
	for _, choiceStr := range d.GetChoicesNames() {
		out += ufmt.Sprintf("[%s](%s) - ", choiceStr, txlink.Call("DaoRespond" + choiceStr, "name", strconv.Itoa(int(d.index))))
	}
	return out
}

// Create a new miner proposal
func newMinerProposal(session *UserSession, who dao.Member, from, to gemType, count uint) *minerProposal {
	session.DAO.globalIndex++
	return &minerProposal{
		author: who,
		votes:  [2]uint{0, 0},
		voters: []dao.Member{},
		executor: &minerExecutor{
			session: session,
			from:    from,
			to:      to,
			count:   count,
		},
		alive: true,
		ended: false,
		index: session.DAO.globalIndex,
	}
}

// The miner dao
type minerDAO struct {
	dao.DAO
	props []dao.Proposal
	globalIndex uint
}

// Get the proposals, also update the dao by removing ended proposals
func (d *minerDAO) GetProposals() []dao.Proposal {
	// newProps := []dao.Proposal{}
	// for _, p := range d.props {
	// 	if !p.IsEnded() {
	// 		newProps = append(newProps, p)
	// 	}
	// }
	// d.props = newProps
	return d.props
}

// Propose a new Proposal
func (d *minerDAO) Propose(prop dao.Proposal) {
	d.props = append(d.props, prop)
}

func newMinerDAO() *minerDAO {
	return &minerDAO{
		dao.NewDAO(),
		[]dao.Proposal{},
		0,
	}
}

// The function for DAO interractions

// pub/ Ask a new DAO request.
func RequestDAO(a, count, from, to string) {
	actor := dao.Member(std.PreviousRealm().Address())
	infos, ok := userSessions[a]
	if !ok {
		panic("invalid session")
	}
	infos.DAO.AssertUserIsMember(actor)

	icount := mustAtoi(count)
	ifrom := mustAtoi(from)
	ito := mustAtoi(to)
	prop := newMinerProposal(infos, actor, gemType(ifrom), gemType(ito), uint(icount))
	infos.DAO.Propose(prop)
}

// pub/ Join a new DAO.
func JoinDAO(what string) {
	actor := dao.Member(std.PreviousRealm().Address())
	infos, ok := userSessions[what]
	if !ok {
		panic("invalid session")
	}
	infos.DAO.AssertUserIsNotMember(actor)
	infos.DAO.AddMember(actor)
}

// pub/ Leave the DAO
func LeaveDAO(what string) {
	actor := dao.Member(std.PreviousRealm().Address())
	infos, ok := userSessions[what]
	if !ok {
		panic("invalid session")
	}
	infos.DAO.AssertUserIsMember(actor)
	if actor.Address() == infos.Owner.Owner() {
		infos.DAO = nil // please god i hope it does garbage collection
		return
	}
	infos.DAO.ExcludeMember(actor)
}

// Rendering
func renderDao(params map[string]string) string {
	user, ok := params["user"]
	if !ok {
		return "user is not specified. [Go back to home](" + gnominersPath + ") to select a user."
	}
	session, ok := userSessions[user]
	if !ok {
		return "this user is not registered. [Go back to home](" + gnominersPath + ") to register."
	}
	if session.DAO == nil {
		return "this user does not opened DAO yet. [Go back to the dashboard](" + gnominersDashboard + "?user=" + user + ")."
	}

	// the content
	var content string
	content += "# DAO of " + session.GetName() + "\n"
	content += "Go back to [DashBoard](" + gnominersDashboard + "?user=" + user + ").\n\n"
	content += "DAO have " + strconv.Itoa(len(session.DAO.GetMembers())) + " members.\n\n"
	content += "## Actions\n"
	content += ufmt.Sprintf("[Join](%s) - [Leave](%s) - [Request](%s)\n\n",
		txlink.Call("JoinDAO", "what", user),
		txlink.Call("LeaveDAO", "what", user),
		txlink.Call("RequestDAO",
			"a", user,
			"count", "10",
			"from", strconv.Itoa(int(gemRock)),
			"to", strconv.Itoa(int(gemQuartz)),
		),

	)
	content += "> It is advised to call `Request` from CLI, for more precise requests. Here, you're converting 10 rocks \u2192 quartz.\n"
	content += "> Also, note that if the owner leaves, the DAO is disbanded.\n\n"

	content += "## Requests\n"
	props := session.DAO.GetProposals()
	if len(props) == 0 {
		content += "Hmmm... It seems that there is no propositions yet.\n"
		return content
	}
	for i := len(props) - 1; i >= 0; i-- {
		p := props[i]
		if p.IsEnded() {
			continue
		}
		content += p.String() + "\n\n"
	}
	content += "\n"

	return content
}
