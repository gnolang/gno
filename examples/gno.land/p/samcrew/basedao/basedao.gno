package basedao

import (
	"chain"
	"chain/runtime"
	"errors"

	"gno.land/p/nt/mux"
	"gno.land/p/samcrew/daocond"
	"gno.land/p/samcrew/daokit"
	"gno.land/p/samcrew/realmid"
)

type daoPublic struct {
	impl *DAOPrivate
}

type SetImplemRaw func(newLocalDAO daokit.DAO, newDAO daokit.DAO)

type RenderFn func(path string, dao *DAOPrivate) string

func (d *daoPublic) Extension(path string) daokit.Extension {
	ext, ok := d.impl.Core.Extensions.Get(path)
	if !ok {
		return nil
	}
	if ext.Info().Private && runtime.CurrentRealm().PkgPath() != d.impl.Realm.PkgPath() {
		panic("attempt to get private extension outside of DAO realm")
	}
	return ext
}

func (d *daoPublic) ExtensionsList() daokit.ExtensionsList {
	return d.impl.Core.Extensions.List()
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

func (d *daoPublic) Render(path string) string {
	return d.impl.Render(path)
}

// DAO is meant for internal realm usage and should not be exposed.
type DAOPrivate struct {
	Core             *daokit.Core
	Members          *MembersStore
	RenderRouter     *mux.Router
	GetProfileString ProfileStringGetter
	Realm            runtime.Realm
	RenderFn         func(path string, dao *DAOPrivate) string
	CallerID         CallerIDFn

	InitialConfig *Config // mostly there in case we need this data during upgrade
}

// Function type that returns the identifier of the current caller.
// Used to identify who is making calls to DAO functions.
type CallerIDFn func() string

type Config struct {
	// Basic DAO information
	Name        string
	Description string
	ImageURI    string

	// Storage handler
	Members *MembersStore

	// Feature toggles
	NoDefaultHandlers  bool // Skips registration of default management actions (add/remove members, etc.)
	NoDefaultRendering bool // Skips setup of default web UI rendering routes
	NoCreationEvent    bool // Skips emitting the DAO creation event

	// Governance configuration
	InitialCondition daocond.Condition // Default condition for all built-in actions, defaults to 60% member majority

	// Profile integration (optional)
	SetProfileString ProfileStringSetter // Function to update profile fields (DisplayName, Bio, Avatar)
	GetProfileString ProfileStringGetter // Function to retrieve profile fields for members

	// Advanced customization hooks
	SetImplemFn       SetImplemRaw      // Function called when DAO implementation changes via governance
	MigrationParamsFn MigrationParamsFn // Function providing parameters for DAO upgrades
	RenderFn          RenderFn          // Rendering function for Gnoweb
	CrossFn           daokit.CrossFn    // Cross-realm communication function for multi-realm DAOs
	CallerID          CallerIDFn        // Custom function to identify the current caller, defaults to realmid.Previous

	// Internal configuration
	PrivateVarName string // Name of the private DAO variable for member querying extensions
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

	if conf.CallerID == nil {
		conf.CallerID = realmid.Previous
	}

	core := daokit.NewCore()
	dao := &DAOPrivate{
		Core:             core,
		Members:          members,
		GetProfileString: conf.GetProfileString,
		Realm:            runtime.CurrentRealm(),
		RenderFn:         conf.RenderFn,
		CallerID:         conf.CallerID,
		InitialConfig:    conf,
	}

	pubdao := &daoPublic{impl: dao}

	dao.Core.Extensions.Set(&membersViewExtension{
		getStore:  func() *MembersStore { return dao.Members },
		queryPath: dao.Realm.PkgPath() + "." + conf.PrivateVarName + ".Members",
	})

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
				Handler:     NewEditProfileHandler(conf.SetProfileString, []string{"DisplayName", "Bio", "Avatar"}),
				Condition:   conf.InitialCondition,
				DisplayName: "Edit Profile",
				Description: "This proposal allows you to edit this DAO profile.",
			})
		}

		if conf.SetImplemFn != nil {
			if conf.CrossFn == nil {
				panic("nil cross function")
			}

			setImplemFn := func(dao daokit.DAO) {
				crossingDAO := daokit.NewCrossing(dao, conf.CrossFn)
				conf.SetImplemFn(dao, crossingDAO)
			}
			setImplemFn(pubdao)

			if conf.MigrationParamsFn == nil {
				conf.MigrationParamsFn = func() []any { return nil }
			}

			dao.Core.Resources.Set(&daokit.Resource{
				Handler:     NewChangeDAOImplementationHandler(dao, setImplemFn, conf.MigrationParamsFn),
				Condition:   conf.InitialCondition,
				DisplayName: "Change DAO Implementation",
				Description: "Change the underlying DAO implementation.",
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

	return pubdao, dao
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
	id := d.CallerID()
	if !d.Members.IsMember(id) {
		panic(errors.New("caller is not a member"))
	}
	return id
}
