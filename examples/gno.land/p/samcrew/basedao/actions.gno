package basedao

import (
	"errors"

	"gno.land/p/moul/md"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/samcrew/daokit"
)

// ADD MEMBER

const ActionAddMemberKind = "gno.land/p/samcrew/basedao.AddMember"

type ActionAddMember struct {
	Address address
	Roles   []string
}

func (a *ActionAddMember) String() string {
	return ufmt.Sprintf("Add member %s with roles %v", a.Address, a.Roles)
}

func NewAddMemberHandler(dao *DAOPrivate) daokit.ActionHandler {
	return daokit.NewActionHandler(ActionAddMemberKind, func(ipayload interface{}) {
		payload, ok := ipayload.(*ActionAddMember)
		if !ok {
			panic(errors.New("invalid payload type"))
		}
		dao.Members.AddMember(payload.Address.String(), payload.Roles)
	})
}

func NewAddMemberAction(payload *ActionAddMember) daokit.Action {
	return daokit.NewAction(ActionAddMemberKind, payload)
}

// REMOVE MEMBER

const ActionRemoveMemberKind = "gno.land/p/samcrew/basedao.RemoveMember"

func NewRemoveMemberHandler(dao *DAOPrivate) daokit.ActionHandler {
	return daokit.NewActionHandler(ActionRemoveMemberKind, func(ipayload interface{}) {
		addr, ok := ipayload.(address)
		if !ok {
			panic(errors.New("invalid payload type"))
		}
		dao.Members.RemoveMember(addr.String())
	})
}

func NewRemoveMemberAction(addr address) daokit.Action {
	return daokit.NewAction(ActionRemoveMemberKind, addr)
}

// ASSIGN ROLE

const ActionAssignRoleKind = "gno.land/p/samcrew/basedao.AssignRole"

type ActionAssignRole struct {
	Address address
	Role    string
}

func (a *ActionAssignRole) String() string {
	return ufmt.Sprintf("Assign role %q to user %s", a.Role, a.Address)
}

func NewAssignRoleHandler(dao *DAOPrivate) daokit.ActionHandler {
	return daokit.NewActionHandler(ActionAssignRoleKind, func(i interface{}) {
		payload, ok := i.(*ActionAssignRole)
		if !ok {
			panic(errors.New("invalid payload type"))
		}
		dao.Members.AddRoleToMember(payload.Address.String(), payload.Role)
	})
}

func NewAssignRoleAction(payload *ActionAssignRole) daokit.Action {
	return daokit.NewAction(ActionAssignRoleKind, payload)
}

// UNASSIGN ROLE

const ActionUnassignRoleKind = "gno.land/p/samcrew/basedao.UnassignRole"

type ActionUnassignRole struct {
	Address address
	Role    string
}

func (a *ActionUnassignRole) String() string {
	return ufmt.Sprintf("Remove role %q from user %s", a.Role, a.Address)
}

func NewUnassignRoleHandler(dao *DAOPrivate) daokit.ActionHandler {
	return daokit.NewActionHandler(ActionUnassignRoleKind, func(i interface{}) {
		payload, ok := i.(*ActionUnassignRole)
		if !ok {
			panic(errors.New("invalid payload type"))
		}
		dao.Members.RemoveRoleFromMember(payload.Address.String(), payload.Role)
	})
}

func NewUnassignRoleAction(payload *ActionUnassignRole) daokit.Action {
	return daokit.NewAction(ActionUnassignRoleKind, payload)
}

// EDIT PROFILE
//
// EditProfile is the only basedao action that crosses into a foreign
// realm (the consumer-supplied profile store). It cannot go through the
// generic daokit.ActionHandler dispatch because that interface is
// cur-free by design — putting rlm on it would be Class-1a (see
// docs/resources/gno-security.md). Instead:
//
//   - editProfileHandler is a concrete basedao-internal type with an
//     UNEXPORTED execute method. No foreign type can satisfy an
//     interface that includes an unexported method, so no hostile
//     implementation can be substituted in.
//   - editProfileStubHandler is registered in daokit.Core.Resources so
//     that proposal-creation can find the action's Condition. Its
//     Execute method panics — if reached, a caller bypassed
//     DAOWrapper.Execute (the rlm-aware path).
//   - DAOWrapper.Execute → DAOPrivate.executeWithRlm uses
//     Core.ExecuteWithDispatch to intercept ActionEditProfileKind and
//     dispatch through editProfileHandler.execute(0, rlm, action),
//     enabling cross2(rlm) instead of a bare cross.

const ActionEditProfileKind = "gno.land/p/samcrew/basedao.EditProfile"

// ActionEditProfile implements daokit.Action directly (rather than being
// wrapped via daokit.NewAction's *genericAction). The rlm-aware
// dispatcher in basedao type-asserts action.(*ActionEditProfile) to
// reach the kv payload; that only works if the proposal carries
// *ActionEditProfile as the Action value, not a generic wrapper around
// it. (genericAction's payload field is unexported in daokit, so basedao
// cannot unwrap.)
type ActionEditProfile struct {
	kv [][2]string
}

func (a *ActionEditProfile) String() string {
	elems := []string{}
	for _, v := range a.kv {
		elems = append(elems, v[0]+": "+v[1])
	}
	return md.BulletList(elems)
}

func (a *ActionEditProfile) Type() string { return ActionEditProfileKind }

func NewEditProfileAction(kv ...[2]string) daokit.Action {
	return &ActionEditProfile{kv: kv}
}

// editProfileHandler is the concrete, rlm-aware EditProfile dispatcher.
// Stored on *DAOPrivate.editProfile (not in Core.Resources), invoked
// from DAOPrivate.executeWithRlm under DAOWrapper.Execute.
type editProfileHandler struct {
	setter        ProfileStringSetter
	allowedFields []string
}

// execute is unexported intentionally — see the file-level EditProfile
// comment. The unexported method name binds the interface signature to
// the basedao package, preventing any foreign type from being
// substituted into basedao's rlm-bearing dispatch path.
func (h *editProfileHandler) execute(_ int, rlm realm, action daokit.Action) {
	a, ok := action.(*ActionEditProfile)
	if !ok {
		panic(errors.New("invalid action type"))
	}
	for _, elem := range a.kv {
		k, v := elem[0], elem[1]
		if len(h.allowedFields) > 0 && !stringSliceContains(h.allowedFields, k) {
			panic(ufmt.Errorf("unauthorized field %q", k))
		}
		h.setter(cross2(rlm), k, v)
	}
}

// editProfileStubHandler is a daokit.ActionHandler registered in
// Core.Resources for ActionEditProfileKind so that Core.Propose's
// Condition lookup succeeds. Its Execute panics — EditProfile MUST be
// dispatched via DAOWrapper.Execute (the rlm-aware path); reaching this
// method means the caller bypassed that contract.
type editProfileStubHandler struct{}

func (editProfileStubHandler) Type() string { return ActionEditProfileKind }

func (editProfileStubHandler) Execute(action daokit.Action) {
	panic(errors.New("ActionEditProfileKind must be dispatched via basedao.DAOWrapper.Execute (rlm-aware path)"))
}

func stringSliceContains(s []string, target string) bool {
	for _, elem := range s {
		if elem == target {
			return true
		}
	}
	return false
}
