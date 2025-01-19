package teams

import (
	"errors"
	"regexp"
	"std"

	"gno.land/p/demo/avl"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrAleadyExist  = errors.New("already exist")
	ErrDoesNotExist = errors.New("does not exist")
)

type AccessController interface {
	CanPerform(member std.Address, msg Msg) bool
}

type Lifecycle interface {
	Init() Action
	Update(team *Team, msg Msg) Action
}

type ITeam interface {
	AccessController
	Lifecycle
}

var teams avl.Tree // std.Address -> Team

type Team struct {
	AccessController
	Lifecycle

	address std.Address
	members avl.Tree // std.Address -> void
}

func (team *Team) Perform(msgs ...Msg) {
	caller := std.GetOrigCaller()
	if !team.IsMember(caller) {
		panic("only member can perform action on team")
	}

	// get actions for the given msgs
	actions := team.getActionsForMsg(msgs...)
	team.performActions(actions...)
}

func (team *Team) AddMember(member std.Address) {
	team.Perform(AddMemberMsg{
		Member: member,
	})
}

func (team *Team) RemoveMember(member std.Address) {
	team.Perform(RemoveMemberMsg{
		Member: member,
	})
}

func (team *Team) HasMember(member std.Address) bool {
	return team.members.Has(member.String())
}

type assertPerformDefaultMsg struct{ assert bool }

func (team *Team) assertPerformDefault() {
	action := team.Lifecycle.Update(team, assertPerformDefaultMsg{})
	msg := action.call(team)
	if assertMsg, ok := msg.(assertPerformDefaultMsg); ok {
		if assertMsg.assert {
			return
		}
	}

	panic(`make sure that team implementation handle team update fallback`)
}

func (team *Team) PerformDefault(msg Msg) Action {
	switch typ := msg.(type) {
	case AddMemberMsg:
		return AddMemberAction(typ.Member)
	case RemoveMemberMsg:
		return RemoveMemberAction(typ.Member)
	case AddPackageMsg: // Do nothing

	case assertPerformDefaultMsg:
		typ.assert = true
		return ActionableMsg(typ)
	}

	return nil
}

func (team *Team) performActions(actions ...Action) {
	var action Action
	for len(actions) > 0 {
		action, actions = actions[0], actions[1:] // shift action
		if action == nil {

			continue // skip empty action
		}

		if msg := action.call(team); msg != nil {
			nextActions := team.getActionsForMsg(msg)
			actions = append(nextActions, actions...)
		}
	}
}

func (team *Team) getActionsForMsg(msgs ...Msg) []Action {
	caller := std.GetOrigCaller()
	actions := make([]Action, 0, len(msgs))
	for _, msg := range msgs {
		if msg == nil {
			continue // skip empty msg
		}

		switch typ := msg.(type) {
		case []Msg:
			subActions := team.getActionsForMsg(typ...)
			actions = append(actions, subActions...)
		default:
			// Assert caller can perform action
			if !team.IsTeamAddress(caller) && !team.AccessController.CanPerform(caller, msg) {
				panic("unauthorized: " + caller.String())
			}

			// Prepare action
			actions = append(actions, team.Lifecycle.Update(team, msg))
		}
	}
	return actions
}

func (team *Team) IsTeamAddress(teamAddr std.Address) bool {
	return teamAddr == team.address
}

func (team *Team) IsMember(member std.Address) bool {
	return member == team.address || team.members.Has(member.String())
}

type unlimitedAC struct{}

func (unlimitedAC) CanPerform(member std.Address, msg Msg) bool {
	return true
}

// realm of the form `xxx.xx/r/<myteam>/home`
var reHomeRealm = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}/r/[a-z0-9_]+/home$`)

func Register(iteam ITeam) *Team {
	caller := std.GetOrigCaller()
	println(caller)
	realm := std.PrevRealm()

	// First lets check the realm is valid

	// check if origin caller is an home path
	if !reHomeRealm.MatchString(realm.PkgPath()) {
		panic("cannot register a team outside an home realm")
	}

	// check if caller is not already registerer as a team
	if teams.Has(caller.String()) {
		panic("team already registered: " + caller)
	}

	// Then initilize the team
	team := &Team{address: caller, Lifecycle: iteam}

	// Assert that team implementation correctly use fallback
	// XXX: do we want this ?
	// > it  assert that caller have a minimal implementation of his team
	team.assertPerformDefault()

	if initAction := iteam.Init(); initAction != nil {
		// init is performed using an unlimited ac
		team.AccessController = unlimitedAC{}
		team.performActions(initAction)
	}

	// once done, apply the provided implementation ac
	team.AccessController = iteam
	teams.Set(caller.String(), team)
	return team
}

func IsRegister(teamAddr std.Address) bool {
	return teams.Has(teamAddr.String())
}
