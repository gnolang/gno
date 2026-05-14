package basedao

import (
	"chain"
	"chain/runtime"
	"errors"

	"gno.land/p/nt/mux/v0"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
)

type daoPublic struct {
	impl *DAOPrivate
}

func (d *daoPublic) Propose(req daokit.ProposalRequest) uint64 {
	return d.impl.Propose(req)
}

func (d *daoPublic) Execute(id uint64) {
	d.impl.Execute(id)
}

func (d *daoPublic) Vote(id uint64, vote daocond.Vote) {
	d.impl.Vote(id, vote)
}

// DAOPrivate is meant for internal realm usage and should not be exposed
type DAOPrivate struct {
	Core             *daokit.Core
	Members          *MembersStore
	RenderRouter     *mux.Router
	GetProfileString ProfileStringGetter
	Realm            runtime.Realm
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

func New(conf *Config) (daokit.DAO, *DAOPrivate) {
	// XXX: emit events from memberstore

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
		Realm:            runtime.CurrentRealm(),
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

	return &daoPublic{impl: dao}, dao
}

func (d *DAOPrivate) Vote(proposalID uint64, vote daocond.Vote) {
	if len(vote) > 16 {
		panic("invalid vote")
	}

	voterID := d.assertCallerIsMember()
	d.Core.Vote(voterID, proposalID, vote)
}

func (d *DAOPrivate) Execute(proposalID uint64) {
	_ = d.assertCallerIsMember()
	d.Core.Execute(proposalID)
}

func (d *DAOPrivate) Propose(req daokit.ProposalRequest) uint64 {
	proposerID := d.assertCallerIsMember()
	return d.Core.Propose(proposerID, req)
}

func (d *DAOPrivate) assertCallerIsMember() string {
	id := runtime.PreviousRealm().Address().String()
	if !d.Members.IsMember(id) {
		panic(errors.New("caller is not a member"))
	}
	return id
}
