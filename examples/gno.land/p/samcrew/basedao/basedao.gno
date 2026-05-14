package basedao

import (
	"chain"
	"chain/runtime"
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
	w.impl.Execute(executor, id)
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
func New(conf *Config) *DAOWrapper {
	// XXX: emit events from memberstore

	members := conf.Members
	if members == nil {
		members = NewMembersStore(nil, nil)
	}

	if conf.GetProfileString == nil {
		panic(errors.New("GetProfileString is required"))
	}

	curRealm := runtime.CurrentRealm()
	core := daokit.NewCore()
	dao := &DAOPrivate{
		Core:             core,
		Members:          members,
		GetProfileString: conf.GetProfileString,
		RealmPkgPath:     curRealm.PkgPath(),
		RealmAddress:     curRealm.Address(),
	}
	dao.initRenderingRouter()

	if !conf.NoDefaultRendering {
		dao.InitDefaultRendering()
	}

	if conf.SetProfileString != nil {
		conf.SetProfileString(cross, "DisplayName", conf.Name)
		conf.SetProfileString(cross, "Bio", conf.Description)
		conf.SetProfileString(cross, "Avatar", conf.ImageURI)
	}

	if !conf.NoDefaultHandlers {
		if conf.InitialCondition == nil {
			conf.InitialCondition = daocond.MembersThreshold(0.6, members.IsMember, members.MembersCount)
		}

		if conf.SetProfileString != nil {
			dao.Core.Resources.Set(&daokit.Resource{
				Handler:   NewEditProfileHandler(conf.SetProfileString, []string{"DisplayName", "Bio", "Avatar"}),
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
func (d *DAOPrivate) Execute(caller address, proposalID uint64) {
	_ = d.assertCallerIsMember(caller)
	d.Core.Execute(proposalID)
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
