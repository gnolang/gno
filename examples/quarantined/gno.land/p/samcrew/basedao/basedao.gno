// Package basedao is a reference DAO implementation built on
// p/samcrew/daokit. Used by the r/samcrew/daodemo example realm.
//
// ─── STATUS: EXPERIMENTAL / INSECURE / INCOMPLETE / COMPLICATED ───
//
// Do not use as the basis for a production DAO without significant
// hardening. The gnoland chain's actual governance is r/gov/dao/v3
// (independent code path; not built on basedao).
//
// Known issues (non-exhaustive):
//
//   - DAOWrapper.Private() returns the raw *DAOPrivate. Its Core,
//     Members, and Resources fields are all exported and mutable.
//     A caller that holds *DAOPrivate can mutate members, swap
//     handlers, or invoke Core.Execute directly with any address,
//     bypassing IsCurrent / member-only / proposal-vote gates. The
//     "DAOPrivate is internal-only" contract is documented but
//     unenforced — anyone exposing *DAOPrivate outside their own
//     realm has handed the DAO away.
//
//   - MembersStore.RemoveMember has a role-bookkeeping bug: the line
//     `role.Members.Remove(role.Name)` should be `Remove(member)`.
//     Removed members continue to count toward Role-keyed quorum
//     math (RoleThreshold / RoleCount conditions) until the role
//     itself is removed.
//
//   - No "at-least-one-member" invariant. RemoveMember accepts going
//     to zero members; the resulting DAO is permanently bricked
//     (IsMember returns false for all callers, all writes panic).
//     Similarly, no "at-least-one-admin" guard if you use a
//     RoleThreshold condition.
//
//   - Address/role-name input validation is missing in handlers.
//     ActionAddMember.Address is not checked for .IsValid(); empty
//     or zero addresses produce phantom voters that can be added
//     and granted roles. Role names accept arbitrary strings
//     including empty / unicode / slashes (the last breaks
//     render-URL construction).
//
//   - EditProfile dispatch uses an rlm-aware path via
//     ExecuteWithDispatch + editProfileStubHandler. If a caller
//     reaches the legacy DAOPrivate.Execute → Core.Execute path
//     for an EditProfile proposal, the stub panics AFTER
//     UpdateStatus mutates Open→Passed, leaving the proposal
//     permanently stuck. Production daodemo doesn't trigger this;
//     tests don't either.
//
//   - See p/samcrew/daokit's package doc for additional issues
//     inherited from the underlying toolkit (re-entrancy via
//     Condition.Eval, stuck-Passed handler-panic state,
//     resource-overwrite mid-flight, etc.)
//
// Architectural complexity notes:
//
//   - The action dispatch path threads through three packages
//     (basedao → daokit Core → registered Handler), with an
//     additional carve-out (DAOPrivate.executeWithRlm +
//     editProfileHandler) for cross-realm dispatch that needs a
//     live cur. Auditing any single action requires understanding
//     all three layers.
//   - DAOWrapper (rlm-typed public surface) and DAOPrivate
//     (address-typed internal) is the Authority-pattern split;
//     the split is correct in intent but Private() escapes it.
//
// Use at your own risk. If you adopt this package, fork it and
// harden the items above before exposing it to untrusted input.
package basedao

import (
	"chain"
	"errors"

	"gno.land/p/nt/mux/v0"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

// DAOWrapper is the public, rlm-validating surface of a basedao DAO. All
// mutating methods take (_ int, rlm realm, ...), validate via
// rlm.IsCurrent(), and derive the principal from rlm.Previous().Address()
// before invoking the private impl's address-typed methods. This is the
// trusted boundary — the daokit.DAO interface itself is dispatch-only and
// never exposed publicly.
//
// Consumers in /r/ should hold `*DAOWrapper`, never `daokit.DAO`.
type DAOWrapper struct {
	impl *DAOPrivate
}

func (w *DAOWrapper) Propose(_ int, rlm realm, req daokit.ProposalRequest) uint64 {
	if !rlm.IsCurrent() {
		panic(errors.New("caller is not a member"))
	}
	proposer := rlm.Previous().Address()
	return w.impl.Propose(proposer, req)
}

func (w *DAOWrapper) Execute(_ int, rlm realm, id uint64) {
	if !rlm.IsCurrent() {
		panic(errors.New("caller is not a member"))
	}
	executor := rlm.Previous().Address()
	// Route through the rlm-aware path: ActionEditProfileKind is dispatched
	// via editProfileHandler.execute(0, rlm, ...) which uses cross(rlm).
	// All other action kinds fall through to generic Core dispatch.
	w.impl.executeWithRlm(0, rlm, executor, id)
}

func (w *DAOWrapper) Vote(_ int, rlm realm, id uint64, vote daocond.Vote) {
	if !rlm.IsCurrent() {
		panic(errors.New("caller is not a member"))
	}
	voter := rlm.Previous().Address()
	w.impl.Vote(voter, id, vote)
}

// InstantExecute proposes, votes Yes, and executes in one call. Each step
// goes through this wrapper's rlm-validated methods, so IsCurrent is
// re-checked at every step (defense-in-depth against drift if the impl is
// later changed to skip per-call validation).
func (w *DAOWrapper) InstantExecute(_ int, rlm realm, req daokit.ProposalRequest) uint64 {
	id := w.Propose(0, rlm, req)
	w.Vote(0, rlm, id, daocond.VoteYes)
	w.Execute(0, rlm, id)
	return id
}

// Render delegates to the private impl's renderer. The impl's render
// surface is read-only (no rlm/cur in any render method).
func (w *DAOWrapper) Render(path string) string {
	return w.impl.Render(path)
}

// Private returns the internal *DAOPrivate. This is intended ONLY for
// the realm that owns the DAOWrapper to wire up custom handlers or
// extend rendering — never expose the result to other realms. Foreign
// code that holds a *DAOPrivate can call its address-typed writers
// directly with any address argument, which would bypass IsCurrent.
func (w *DAOWrapper) Private() *DAOPrivate {
	return w.impl
}

// DAOPrivate is meant for internal realm usage and should not be exposed.
//
// SECURITY: writer methods (Propose, Vote, Execute) are address-typed and
// trust the supplied caller. They DO NOT validate IsCurrent — that
// validation lives in DAOWrapper. Holding *DAOPrivate and calling its
// methods with a forged caller address is a Class-2 designation-forgery.
// Owning realms must keep their *DAOPrivate in an unexported package var
// and never return it from any exported function. Use DAOWrapper.Private()
// only at construction time to register custom action handlers.
type DAOPrivate struct {
	Core             *daokit.Core
	Members          *MembersStore
	RenderRouter     *mux.Router
	GetProfileString ProfileStringGetter
	// RealmPkgPath / RealmAddress are captured at construction time as
	// plain string/address values rather than a runtime.Realm struct.
	// Storing a runtime.Realm in a struct field would be a parallel
	// anti-pattern to storing a `cur realm` (capability token in a
	// struct). PkgPath/Address are stable identity strings, safe to
	// persist.
	RealmPkgPath string
	RealmAddress address
	// editProfile carries the rlm-aware dispatcher for
	// ActionEditProfileKind. nil if Config.SetProfileString was nil.
	// See actions.gno for the rationale (concrete type + unexported
	// method closes Class-1a; basedao routes around the generic
	// daokit.ActionHandler interface for this one action).
	editProfile *editProfileHandler
}

type Config struct {
	Name               string
	Description        string
	ImageURI           string
	Members            *MembersStore
	NoDefaultHandlers  bool
	NoDefaultRendering bool
	InitialCondition   daocond.Condition
	SetProfileString   ProfileStringSetter
	GetProfileString   ProfileStringGetter
	NoCreationEvent    bool
}

type ProfileStringSetter func(cur realm, field string, value string) bool
type ProfileStringGetter func(addr address, field string, def string) string

const EventBaseDAOCreated = "BaseDAOCreated"

// New constructs a basedao DAO and returns its public wrapper. The
// returned *DAOWrapper is the canonical type for /r/ consumers to hold;
// it exposes rlm-validating Propose / Vote / Execute / InstantExecute /
// Render methods, and a .Private() accessor for the owning realm to wire
// custom action handlers at init time.
//
// rlm is threaded into the optional `conf.SetProfileString` callback via
// cross(rlm). /p/ production code cannot declare crossing functions, so
// rlm is taken as a non-first parameter — callers pass `cur` directly.
func New(conf *Config, rlm realm) *DAOWrapper {
	// XXX: emit events from memberstore

	// IsCurrent gate: rlm must be the caller's live cur. Otherwise
	// RealmPkgPath/RealmAddress (captured below into a long-lived
	// struct) would be set from a stale or sibling-frame value, and
	// the cross(rlm) calls further down would panic anyway when
	// SetProfileString is non-nil. Catch the misuse up-front so the
	// `SetProfileString == nil` path is also guarded.
	if !rlm.IsCurrent() {
		panic("basedao.New: rlm is not the caller's live cur")
	}

	members := conf.Members
	if members == nil {
		members = NewMembersStore(nil, nil)
	}

	if conf.GetProfileString == nil {
		panic(errors.New("GetProfileString is required"))
	}

	core := daokit.NewCore()
	dao := &DAOPrivate{
		Core:             core,
		Members:          members,
		GetProfileString: conf.GetProfileString,
		RealmPkgPath:     rlm.PkgPath(),
		RealmAddress:     rlm.Address(),
	}
	dao.initRenderingRouter()

	if !conf.NoDefaultRendering {
		dao.InitDefaultRendering()
	}

	if conf.SetProfileString != nil {
		conf.SetProfileString(cross(rlm), "DisplayName", conf.Name)
		conf.SetProfileString(cross(rlm), "Bio", conf.Description)
		conf.SetProfileString(cross(rlm), "Avatar", conf.ImageURI)
	}

	if !conf.NoDefaultHandlers {
		if conf.InitialCondition == nil {
			conf.InitialCondition = daocond.MembersThreshold(0.6, members.IsMember, members.MembersCount)
		}

		if conf.SetProfileString != nil {
			// editProfile is the rlm-aware dispatcher used by
			// DAOPrivate.executeWithRlm. The Resource entry below uses a
			// panicking stub (editProfileStubHandler) so Core.Propose's
			// Condition lookup succeeds; the stub is unreachable on the
			// production execute path (DAOWrapper.Execute routes via
			// executeWithRlm with an override that intercepts EditProfile).
			dao.editProfile = &editProfileHandler{
				setter:        conf.SetProfileString,
				allowedFields: []string{"DisplayName", "Bio", "Avatar"},
			}
			dao.Core.Resources.Set(&daokit.Resource{
				Handler:   editProfileStubHandler{},
				Condition: conf.InitialCondition,
			})
		}

		defaultResources := []daokit.Resource{
			{
				Handler:     NewAddMemberHandler(dao),
				Condition:   conf.InitialCondition,
				DisplayName: "Add Member",
				Description: "This proposal allows you to add a new member to the DAO.",
			},
			{
				Handler:     NewRemoveMemberHandler(dao),
				Condition:   conf.InitialCondition,
				DisplayName: "Remove Member",
				Description: "This proposal allows you to remove a member from the DAO.",
			},
			{
				Handler:     NewAssignRoleHandler(dao),
				Condition:   conf.InitialCondition,
				DisplayName: "Assign Role",
				Description: "This proposal allows you to assign a role to a member.",
			},
			{
				Handler:     NewUnassignRoleHandler(dao),
				Condition:   conf.InitialCondition,
				DisplayName: "Unassign Role",
				Description: "This proposal allows you to unassign a role from a member.",
			},
		}
		// register management handlers
		for _, resource := range defaultResources {
			dao.Core.Resources.Set(&resource)
		}

	}

	if !conf.NoCreationEvent {
		chain.Emit(EventBaseDAOCreated)
	}

	return &DAOWrapper{impl: dao}
}

// Vote on a proposal. caller must be a current DAO member; the wrapper
// (DAOWrapper.Vote) derives this from rlm.Previous().Address() under
// IsCurrent. Direct callers of DAOPrivate.Vote forge the caller at their
// own risk — see the type-level security note.
func (d *DAOPrivate) Vote(caller address, proposalID uint64, vote daocond.Vote) {
	if len(vote) > 16 {
		panic("invalid vote")
	}
	voterID := d.assertCallerIsMember(caller)
	d.Core.Vote(voterID, proposalID, vote)
}

// Execute a proposal. See Vote for caller semantics.
//
// Address-typed legacy path (preserves the daokit.DAO interface contract
// of no rlm crossing the dispatch boundary). EditProfile proposals
// reaching this method will panic via the registered editProfileStubHandler
// — they must be dispatched via DAOPrivate.executeWithRlm (which
// DAOWrapper.Execute calls). See daokit.gno ExecuteWithDispatch CAVEAT.
func (d *DAOPrivate) Execute(caller address, proposalID uint64) {
	_ = d.assertCallerIsMember(caller)
	d.Core.Execute(proposalID)
}

// executeWithRlm is the rlm-aware execute path. Called by DAOWrapper.Execute
// after IsCurrent has been validated. Asserts caller membership (same as
// DAOPrivate.Execute), then dispatches through Core.ExecuteWithDispatch
// with an override that routes ActionEditProfileKind into the
// editProfile concrete handler. All other action types fall through to
// generic Core dispatch unchanged.
func (d *DAOPrivate) executeWithRlm(_ int, rlm realm, caller address, proposalID uint64) {
	_ = d.assertCallerIsMember(caller)
	d.Core.ExecuteWithDispatch(proposalID, func(action daokit.Action) bool {
		if action.Type() == ActionEditProfileKind && d.editProfile != nil {
			d.editProfile.execute(0, rlm, action)
			return true
		}
		return false
	})
}

// Propose a proposal. See Vote for caller semantics.
func (d *DAOPrivate) Propose(caller address, req daokit.ProposalRequest) uint64 {
	proposerID := d.assertCallerIsMember(caller)
	return d.Core.Propose(proposerID, req)
}

// assertCallerIsMember returns the caller as a string after verifying
// it's a current DAO member.
func (d *DAOPrivate) assertCallerIsMember(caller address) string {
	id := caller.String()
	if !d.Members.IsMember(id) {
		panic(errors.New("caller is not a member"))
	}
	return id
}
