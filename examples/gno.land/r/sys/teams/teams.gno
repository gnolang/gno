package teams

import (
	"errors"
	"regexp"
	"std"

	"gno.land/p/demo/avl"
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrAleadyExist        = errors.New("already exist")
	ErrDoesNotExist       = errors.New("does not exist")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrActionDoesNotExist = errors.New("action does not exist")
)

var teams avl.Tree // std.Address -> Team

type AccessController interface {
	CanPerform(member std.Address, action ActionMsg) bool
}

type Team struct {
	std.Address

	ac      AccessController
	members avl.Tree // std.Address -> void
	actions avl.Tree // append only Action -> ActionDescription
}

func (team *Team) Perform(action ActionName, resource ActionMsg) {
	actionDesc, ok := team.actions.Get(string(action))
	if !ok {
		panic(ErrActionDoesNotExist)
	}

	caller := std.GetOrigCaller()
	if !team.ac.CanPerform(caller, action, resource) {
		panic(ErrUnauthorized)
	}

	handler(team, resource)
}

func (team *Team) RegisterAction(action ActionDescription) {
	team.Perform(AddAction, action)
}

func (team *Team) OverrideAction(previous ActionName, new ActionDescription) {
	previousAction, ok := team.getAction(previous)
	if ok {
		panic("previous action doesn't exist")
	}

	if previousAction.Action == new.Action {
		panic("cannot override a function with the same name")
	}

	team.Perform(SetActionStatus, ActionStatusParam{
		Action: previous,
		Status: Forbidden,
	})
	team.Perform(AddAction, new)
}

func (team *Team) getAction(action ActionName) (ad ActionDescription, ok bool) {
	var val interface{}
	if val, ok = team.actions.Get(string(action)); ok {
		ad = val.(ActionDescription)
	}
	return ad, ok
}

func Get(team std.Address) *Team {
	if t, ok := teams.Get(team.String()); ok {
		return t.(*Team)
	}
	return nil
}

// realm of the form `xxx.xx/r/<myteam>/home`
var reHomeRealm = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}/r/[a-z0-9_]+/home$`)

type unlimitedAC struct{}

func (unlimitedAC) CanPerform(member std.Address, action ActionName, msg ActionMsg) bool {
	return true
}

func Register(ac AccessController, init func(*Team)) *Team {
	caller := std.GetOrigCaller()
	realm := std.PrevRealm()

	// check if origin caller is an home path
	if !reHomeRealm.MatchString(realm.PkgPath()) {
		panic("cannot register a team outside an home realm")
	}

	// check if caller is not already registerer
	if teams.Has(caller.String()) {
		panic("team already registered: " + caller)
	}

	// Init
	team := &Team{ac: &unlimitedAC{}}
	init(team)

	// Register user controller
	team.ac = ac

	// Add the team
	teams.Set(caller.String(), team)

	return team
}

func IsRegister(team std.Address) bool {
	return teams.Has(team.String())
}
