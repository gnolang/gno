package teams

import "std"

type ActionMsg interface{}

type AddMemberMsg struct {
	Member std.Address
}

type RemoveMemberMsg struct {
	Member std.Address
}

type AddPackageMsg struct {
	Path string
}

// Core actions
const (
	AddMember           ActionName = "add_member"
	RemoveMember                   = "remove_member"
	AddPackage                     = "add_pkg"
	SetActionVisibility            = "set_action_status"
	AddAction                      = "add_action"    // Allowed ?
	UpdateAccessControl            = "update_implem" // Allowed ?
)

func (a ActionName) String() string {
	return string(a)
}

type ActionHandler func(t *Team, msg ActionMsg)

type ActionDescription struct {
	Action      ActionName
	Description string
	Private     bool
	Deprecated  bool
	handler     ActionHandler
}

var actionDescriptions = []ActionDescription{
	{
		Action:      AddMember,
		Description: "Adds a new member to the team.",
		handler:     handlerAddMember,
	},
	{
		Action:      RemoveMember,
		Description: "Removes an existing member from the team.",
		handler:     handlerRemoveMember,
	},
	{
		Action:      AddPackage,
		Description: "Adds a new package to the team's resources.",
		handler:     nil, // XXX
	},
	{
		Action:      AddAction,
		Description: "Registers a new action for the team.",
		handler:     handlerRegisterAction,
	},
	{
		Action:      UpdateAccessControl,
		Description: "Updates the access control interface of a team action.",
		handler:     nil, // XXX
	},
}

type SetActionVisibilityMsg struct {
	ActionName
	Private bool
}

func handlerSetActionVisibility(team *Team, msg ActionMsg) {
	param, ok := msg.(SetActionVisibilityMsg)
	if !ok {
		panic("invalid arguments")
	}

	action, ok := team.getAction(param.ActionName)
	if !ok {
		panic("action doesn't exsit")
	}

	// Set new status
	action.Private = param.Private
	team.actions.Set(string(action.Action), action)
}

func handlerRegisterAction(team *Team, msg ActionMsg) {
	action, ok := msg.(ActionDescription)
	if !ok {
		panic("invalid arguments")
	}

	if team.actions.Has(string(action.Action)) {
		panic(ErrAleadyExist)
	}

	team.actions.Set(string(action.Action), action)
}

func handlerUpdateImplementation(team *Team, msg ActionMsg) {
	action, ok := msg.(ActionDescription)
	if ok {
		panic("invalid arguments")
	}

	if team.actions.Has(string(action.Action)) {
		panic(ErrAleadyExist)
	}

	team.actions.Set(string(action.Action), action)
}

func handlerAddMember(team *Team, msg ActionMsg) {
	member, ok := msg.(std.Address)
	if !ok {
		panic("invalid arguments")
	}

	if team.members.Has(member.String()) {
		panic(ErrAleadyExist)
	}
	team.members.Set(member.String(), struct{}{})
}

// func handlerAddPackage(team *Team, msg ActionMsg) {}

func handlerRemoveMember(team *Team, msg ActionMsg) {
	member, ok := msg.(std.Address)
	if !ok {
		panic("invalid arguments")
	}

	if !team.members.Has(member.String()) {
		panic(ErrDoesNotExist)
	}

	team.members.Remove(member.String())
}
