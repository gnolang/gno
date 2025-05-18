package gnominers

import (
	"errors"
	"math"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
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
	count   uint         // How many gems you use in from
	from    gemType      // The gem you want to give
	to      gemType      // The gem you want to convert
	success uint         // If successful
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
	alive    bool           // If the poll is alive
	index    uint           // the proposal index
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

func DaoResponse(a, name string, choice dao.ProposalChoice) {
	actor := dao.Member(std.PreviousRealm().Address())
	infos, ok := userSessions[a]
	if !ok {
		panic("invalid session")
	}
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

func DaoRespondYes(a, name string) {
	DaoResponse(a, name, DaoYes)
}

func DaoRespondNo(a, name string) {
	DaoResponse(a, name, DaoNo)
}

func (d minerProposal) String() string {
	if !d.alive {
		return ""
	}
	if d.ended {
		return md.Blockquote(ufmt.Sprintf("#%d results with %d %s.", d.index, d.executor.success, upperFirstLetter(gemNames[d.executor.to])))
	}

	// header
	t := mdtable.Table{
		Headers: []string{generateCoolName(d.author.Address()) + "'s proposal n°" + strconv.Itoa(int(d.index))},
	}

	// first line (chance + convertion data)
	var s, quote, out string
	if d.executor.count > 1 {
		s = "s"
	}
	if d.executor.from < d.executor.to {
		quote = ufmt.Sprintf(" (%.2f%% chance)", merge(float64(d.executor.to-d.executor.from), d.executor.count)*100.)
	} else {
		quote = ufmt.Sprintf(" (between 0 and %d)", int(unMerge(float64(d.executor.from-d.executor.to), d.executor.count)))
	}
	t.Append([]string{ufmt.Sprintf("Should we convert %d **%s%s** into **%s**%s ?",
		d.executor.count,
		upperFirstLetter(gemNames[d.executor.from]),
		s,
		upperFirstLetter(gemNames[d.executor.to]),
		quote,
	)})

	// second line, votes
	t.Append([]string{strconv.Itoa(int(d.HowManyVoters())) + " votes"})

	// third line, answers
	for k, choiceStr := range d.GetChoicesNames() {
		if k == dao.ProposalChoice(0) {
			out += " - "
		}
		out += md.Link(choiceStr, txlink.Call("DaoRespond"+choiceStr, "a", d.executor.session.Owner.Owner().String(), "name", strconv.Itoa(int(d.index))))
	}
	t.Append([]string{out})
	return t.String()
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
	props       []dao.Proposal
	globalIndex uint
}

func (d *minerDAO) GetProposals() []dao.Proposal {
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

func generateRequestLink(params map[string]string, user string) (string, error) {
	req, ok := params["request"]
	if !ok {
		return "", nil
	}
	list := strings.Split(req, ",")
	if len(list) != 3 {
		return "", errors.New("invalid request")
	}
	for _, v := range list {
		if _, err := strconv.Atoi(v); err != nil {
			return "", err
		}
	}
	return txlink.Call("RequestDAO",
		"a", user,
		"count", list[0],
		"from", list[1],
		"to", list[2],
	), nil
}

// Rendering
func renderDao(params map[string]string) string {
	user, ok := params["user"]
	if !ok {
		return "user is not specified. " + md.Link("Go back to home", gnominersPath) + " to select a user."
	}
	session, ok := userSessions[user]
	if !ok {
		return "this user is not registered. " + md.Link("Go back to home", gnominersPath) + " to register."
	}
	if session.DAO == nil {
		return "this user does not opened DAO yet. " + md.Link("Go back to the dashboard", gnominersDashboard + "?user=" + user) + "."
	}

	// The actions and main content
	var content string
	content += md.H1("DAO of " + session.GetName())
	content += "Go back to " + md.Link("DashBoard", gnominersDashboard + "?user=" + user) + "." + br
	content += "DAO have " + strconv.Itoa(len(session.DAO.GetMembers())) + " members." + br

	content += "<gno-columns>" + br
	content += md.H2("Actions")
	lnk, err := generateRequestLink(params, user)
	if err != nil {
		return "error: " + err.Error() + md.Link("Go back to the dashboard", gnominersDashboard + "?user=" + user)
	}
	content += md.Link("Join", txlink.Call("JoinDAO", "what", user)) + " - " + md.Link("Leave", txlink.Call("LeaveDAO", "what", user))
	if lnk == "" {
		content += " - You can also " + md.Link("Generate A Request", gnominersDAORequest + "?user=" + user) + br
		x := "It is advised to generate a request by click on the link above.\nAlso, note that if the owner leaves, the DAO is disbanded"
		content += md.Blockquote(x) + br
	} else {
		content += md.H3(md.Link("Submit your Request", lnk))
	}
 	content += br

	// The requests
	content += "|||" + br
	content += md.H2("Requests")
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
	content += "</gno-columns>" + br

	return content
}
