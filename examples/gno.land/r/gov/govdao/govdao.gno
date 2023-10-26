package govdao

import (
	"errors"
	"std"
	"strconv"
)

func main() {
	govDAO := GovDAO{}

	worxDAO := BaseDAO{name: "WorxDAO"}
	evaluationDAO := BaseDAO{name: "EvaluationDAO", comment: ""}
	engineeringDAO := EngineeringDAO{BaseDAO{name: "EngineeringDAO", comment: "responsible for high-level architecture"}}
	grantsDAO := EngineeringDAO{BaseDAO{name: "GrantsDAO", comment: "manages grants"}}
	projectsDAO := EngineeringDAO{BaseDAO{name: "ProjectsDAO"}}
	growthDAO := ProjectDAO{BaseDAO{name: "GrowthDAO"}}
	marketingDAO := ProjectDAO{BaseDAO{name: "MarketingDAO"}}
	legalDAO := ProjectDAO{BaseDAO{name: "LegalDAO"}}
	savethecosmosDAO := ProjectDAO{BaseDAO{name: "SaveTheCosmosDAO"}}
	dxDAO := ProjectDAO{BaseDAO{name: "dxDAO", comment: "DevX team"}}
	opsDAO := ProjectDAO{BaseDAO{name: "OpsDAO"}}
	worldDAO := LocalCommunityDAO{BaseDAO{name: "WorldDAO", comment: "local communities"}}
	{
		usa := worldDAO.AddChild("USA")
		california := usa.AddChild("California")
		sf := california.AddChild("SanFrancisco")
		palo := california.AddChild("PaloAlto")
		california.AddChild("...")
		europe := worldDAO.AddChild("Europe")
		fr := europe.AddChild("France")
		paris := fr.AddChild("Paris")
		asia := worldDAO.AddChild("Asia")
		worldDAO.AddChild("...")
		// ...
	}

	// XXX: use orgsdao to show conflict of interests etc
	orgsDAO := OrgDAO{BaseDAO{name: "OrgsDAO"}}
	orgsDAO.children = []DAO{
		&OrgDAO{BaseDAO{name: "AiB"}},
		&OrgDAO{BaseDAO{name: "NT"}},
		&OrgDAO{BaseDAO{name: "..."}},
	}

	// govdao tiers
	tier0 := Tier{idx: 0}
	tier1 := Tier{idx: 1}
	tier2 := Tier{idx: 2}
	govDAO.tiers = []*Tier{&tier0, &tier1, &tier2}
	govDAO.tiers[0].parent = nil
	govDAO.tiers[0].child = govDAO.tiers[1]
	govDAO.tiers[1].parent = govDAO.tiers[0]
	govDAO.tiers[1].child = govDAO.tiers[2]
	govDAO.tiers[2].parent = govDAO.tiers[1]
	govDAO.tiers[2].child = nil

	govDAO.children = []DAO{&worxDAO, &worldDAO, &orgsDAO}
	worxDAO.children = []DAO{&evaluationDAO, &engineeringDAO, &projectsDAO, &opsDAO}
	opsDAO.children = []DAO{&marketingDAO, &legalDAO}
	gameOfRealmsDAO := ProjectDAO{BaseDAO{name: "GameOfRealmsDAO"}}
	projectsDAO.children = []DAO{&savethecosmosDAO, &growthDAO, &gameOfRealmsDAO}
	engTm2DAO := EngineeringDAO{BaseDAO{
		name: "tm2", children: []DAO{
			&BaseDAO{name: "amino"},
			&BaseDAO{name: "bft"},
		},
	}}
	engGnovmDAO := EngineeringDAO{BaseDAO{name: "gnovm"}}
	//&EngineeringDAO{BaseDAO{name: "gnolang"}},
	//&EngineeringDAO{BaseDAO{name: "stdlibs"}},
	engGnolandDAO := EngineeringDAO{BaseDAO{name: "gno.land"}}

	engineeringDAO.children = []DAO{&engTm2DAO, &engGnovmDAO, &engGnolandDAO}
	grantsDAO.children = []DAO{
		&dxDAO,
		&ProjectDAO{BaseDAO{name: "onblocDAO"}},
		&ProjectDAO{BaseDAO{name: "bertyDAO"}},
		&ProjectDAO{BaseDAO{name: "teritoriDAO"}},
	}
	growthDAO.children = []DAO{
		&grantsDAO,
		&ProjectDAO{BaseDAO{name: "student-program"}},
		&ProjectDAO{BaseDAO{name: "builders-program"}},
		&ProjectDAO{BaseDAO{name: "tinkerers-program"}},
	}
	// growthDAO.teams = []Team{marketingTeam, devrelTeam}

	var (
		manfred       = &AccountPerson{NotimplementedActor{}, "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq"}
		jae           = &NoopPerson{NotimplementedActor{}, "jae"}
		guilhem       = &NoopPerson{NotimplementedActor{}, "guilhem"}
		morgan        = &NoopPerson{NotimplementedActor{}, "morgan"}
		milos         = &NoopPerson{NotimplementedActor{}, "milos"}
		petar         = &NoopPerson{NotimplementedActor{}, "petar"}
		antonio       = &NoopPerson{NotimplementedActor{}, "antonio"}
		marc          = &NoopPerson{NotimplementedActor{}, "marc"}
		hariom        = &NoopPerson{NotimplementedActor{}, "hariom"}
		ilker         = &NoopPerson{NotimplementedActor{}, "ilker"}
		adriana       = &NoopPerson{NotimplementedActor{}, "adriana"}
		alex          = &NoopPerson{NotimplementedActor{}, "alex"}
		alexis        = &NoopPerson{NotimplementedActor{}, "alexis"}
		thomas        = &NoopPerson{NotimplementedActor{}, "thomas"}
		michelle      = &NoopPerson{NotimplementedActor{}, "michelle"}
		johnny        = &NoopPerson{NotimplementedActor{}, "johnny"}
		albert        = &NoopPerson{NotimplementedActor{}, "albert"}
		ray           = &NoopPerson{NotimplementedActor{}, "ray"}
		dylan         = &NoopPerson{NotimplementedActor{}, "dylan"}
		valeh         = &NoopPerson{NotimplementedActor{}, "valeh"}
		techStaff     = TeamPerson{NotimplementedActor{}, []Person{manfred, jae, guilhem, morgan, milos, petar, antonio, marc, hariom}}
		coreContribs  = TeamPerson{NotimplementedActor{}, []Person{}}
		marketingTeam = TeamPerson{NotimplementedActor{}, []Person{}}
		growthTeam    = TeamPerson{}
		techCommittee = TeamPerson{NotimplementedActor{}, []Person{manfred, jae, milos, ray}}
	)

	println(govDAO.Directory(""))
}

type NoopPerson struct {
	Actor
	name string
}

func (p NoopPerson) String() string { return p.name }

type AccountPerson struct {
	Actor
	addr std.Address
}

func (p AccountPerson) String() string { return p.addr.String() }

type TeamPerson struct {
	Actor
	members []Person
}

func (p TeamPerson) String() string {
	output := ""
	for i, m := range p.members {
		if i > 0 {
			output += "+"
		}
		output == m.String()
	}
	return output
}

// Person can be a person, a team, an organization, a country, an object?
type Person interface {
	Actor
	String() string
}

// Actor is a Person who can make actions.
type Actor interface {
	Receive(Action) error
}

type NoopActor struct{}

func (a *NoopActor) Receive(Action) error { return nil }

type NotimplementedActor struct{}

func (a NotimplementedActor) Receive(Action) error { return errors.New("not implemented actor") }

// Action defines rules and action.
type Action interface {
	Verify(Actor) error
	Execute(Actor) error
}

type GovDAO struct {
	BaseDAO
	tiers []*Tier
}

type OrgDAO struct {
	BaseDAO
}

// XXX: implement
type ForkableDAO interface {
	Fork()
}

type DAO interface {
	Directory(prefix string) string
	Parent() DAO
	HasChild(DAO) bool
	Owner() Person
	Facilitator() Person
}

func (d GovDAO) Parent() DAO {
	return nil // top-level
}

func (d GovDAO) HasChild(c DAO) bool {
	for _, child := range d.children {
		if child == c {
			return true
		}
	}
	return false
}

func (d GovDAO) Directory(prefix string) string {
	output := prefix + "- GovDAO\n"

	// tiers
	output += prefix + "  - Tiers\n"
	for _, tier := range d.tiers {
		output += tier.Directory(prefix + "    ")
	}

	// children
	for _, child := range d.children {
		// if child.Parent() != d {
		// 	output += prefix + "  <!-- WARN: child claims that we're not their parent:-->\n"
		// }
		output += child.Directory(prefix + "  ")
	}
	return output
}

var _ DAO = &GovDAO{}

type Tier struct {
	idx    int
	parent *Tier
	child  *Tier
}

func (t Tier) Owner() Person       { return nil } // not relevant for tiers
func (t Tier) Facilitator() Person { return nil } // not relevant for tiers

// func (t *Tier) Promote()
// func (t *Tier) Demote()
func (t Tier) Directory(prefix string) string {
	output := prefix + "- tier#" + strconv.Itoa(t.idx) + "\n"
	return output
}

func (t Tier) Parent() DAO {
	return t.parent
}

func (t Tier) HasChild(c DAO) bool {
	return t.child == c
}

type EngineeringDAO struct {
	BaseDAO
}

type ProjectDAO struct {
	BaseDAO
}

type LocalCommunityDAO struct {
	BaseDAO
}

func (d *LocalCommunityDAO) AddChild(name string) *LocalCommunityDAO {
	child := LocalCommunityDAO{BaseDAO{name: name}}
	d.BaseDAO.children = append(d.children, &child)
	// child.parent = d
	return &child
}

type BaseDAO struct {
	name        string
	children    []DAO
	parent      DAO
	comment     string
	owner       Person
	facilitator Person
}

func (d BaseDAO) Owner() Person       { return d.owner }
func (d BaseDAO) Facilitator() Person { return d.facilitator }

func (d BaseDAO) Directory(prefix string) string {
	output := prefix + "- " + d.name
	if d.comment != "" {
		output += " // " + d.comment
	}
	output += "\n"
	childPrefix := prefix + "  "

	for _, child := range d.children {
		// BUG: https://github.com/gnolang/gno/issues/1291
		//_ = child.Parent()
		//parent := child.Parent()
		//if parent != d {
		//}
		//if child.Parent() != d {
		// output += prefix + "  <!-- WARN9: child claims that we're not their parent:-->\n"
		//output += "TTT"
		//}
		output += child.Directory(childPrefix)
	}

	return output
}

func (d BaseDAO) Parent() DAO {
	return d.parent
}

func (d BaseDAO) HasChild(c DAO) bool {
	for _, child := range d.children {
		if child == c {
			return true
		}
	}
	return false
}

/*
// Directory: play with hierarchy by verifying reciprocal permissionless claims
// XXX: move to p/... and make API super short.
type DirectoryEntry interface {
	Parent() DirectoryEntry
	HasChild(DirectoryEntry) bool
}

func DirectorySafeParent(e DirectoryEntry) DirectoryEntry {
	if e == nil {
		return nil
	}
	parent := e.Parent()
	agreement := parent.HasChild(e)
	if agreement {
		return parent
	}
	return nil
}

func DirectoryHasChild(e, c DirectoryEntry) bool {
	if e == nil || c == nil {
		return false
	}
	if !e.HasChild(c) {
		return false
	}
	return c.Parent() == e
}

// HasRecursiveChild
// HasRecursiveParent
   // recursiveTreeWithoutDuplicates (once)
*/

// func (d *GovDAO) Tiers() []Tier
//
// func (d *GovDAO) Directory(level int) string {
// 	for _, tier := range d.tiers {
// 		prefix(level*2+1, tier.Directory())
// 	}
// }
//

// DIRECTORY
// var head Project

// Node is a safe object.
// type Node interface {
// 	SetAssociation(association)
// 	Directory() string
// 	Project // embeds a project
// }

// helpers
// func GetParent()
// func SetParent(node, parent)
// func AddChild(node, child)
// func RemoveChild(node, child)

// - gov/directory        -> explore DAOs and Actors in the network
// - gov/govdao           -> top-level DAO node with tiered members
// - gov/types,interfaces ->
// 	- worx/profiles        -> portfolios built around contributions, endorsements, fact oracles and personal presentations
// - worx/evaluation      -> contributions management
// - worx/dao
//
// - dxteam/dao
// - berty/dao
// - teritori/dao
// - onbloc/dao

// GOVTYPES

// type Project interface {
// 	Children() []Project
// }

// Actor XXX
// Actor can represent a blockchain account (std.Address), a realm (PkgPath), or a go object.
// type Actor interface {
// 	Do(Action)
// 	Verify()
// }

// type Action interface{}

// GOVTPL

// - permissionless: anyone can create a project or an actor; top down adding children
// - an object COULD be Project AND Actor

// func NewTeamAnyActor() Actor                 {}
// func NewRealmActor(pkgpath string) Actor     {} // realm
// func NewAccountActor(addr std.Address) Actor {}
// func NewGoObjActor(impl interface{}) Actor   {} // it is advised to create its own implementation matching a specific interface; human, multisig, oracles
// func NewGovProposal(Action) Proposal         {}

// type CommonProject struct {
// 	Forkable
// }

// func (p *CommonProject) Owners() []Actor
// func (p *CommonProject) Facilitators() []Actor

// var directory Node

// govdao
//   worxdao -> autonomous but also mandated by govdao; HR; "paying salaries"
//   anotherchain

// func init() {
// 	var govdao Node
// 	var personJae Node
// 	var personManfred Node
// 	var worxDAO Node
// 	var companyNT Node // distribute treasury to employees
// 	var chainCosmos Node
// 	var projTM2 Node
// 	var projTM2Amino Node
// 	var projGnovm Node
// 	var projGnoland Node
//
// 	// govdao.AttachChild(worxDAO)
// 	// govdao.Proposal(addNewChild)
// }

// type Contribution struct {
// 	Links []string
// }

// type Contributor struct {
// 	Tier  int
// 	Score int
// }

// worxdao -> HR => bonus, hire, investment, fire, promote, demote,
// earn points/money to review

// worxDAO -> bucket
// levels are managing levels below; they are filtering, reviewing, and score
// proposal ->

// tokens ^worx
// tokens $gnot
// tokens ^gov

// type Directory interface {
// 	Node // embeds a Node
// 	// ...
// }
//
// type EvaluationDAO struct{}
//
// type Manfred struct{}
//
// type NT struct{}

// XXX: switch to Book
// split
// multiple governance; one of them is following levels, not all of them

// 70% contributors
// 20% validators
// 5% contracts
// 5% holders / freezer

// var (
// 	projects []Project
// 	teams    []Team
// 	people   []Person
//         )

// type Project interface {
// 	owner   Team
// 	members Team
// }
//
// type Team struct {
// 	people []Person
// }
//
// type Person struct {
// 	addr std.Address
// }

// worxdao -> distribute work, money
// govdao -> own everything
// evaluation dao 1 -> review contributions

// fork
// parent

// govdao (owned, by top tiers)
//   gno.land
//   gnovm
//   tm2
//    amino

// level1 -> govdao
// level2 -> evelaution dao

// level7
// level6
// level5
// level4
// level3
// level2
// level1

// WorxDAO
// XXX(manfred): please, let me finish a v0, then we can iterate
// highest level == highest responsibility
// teams are responsible for components they don't owne
// flag : realm maintainers VS facilitators
// teams
// committee of trustees to create the directory
// each directory is a name, has a parent and have groups
// homepage team - blocks aggregating events
// XXX: TODO

// --------

// # Directory
//
// * gno.land (owned by group)
// *
// 	* gnovm
// * gnolang (language)
// * gnovm
// - current challenges / concerns / issues
// * tm2
// * amino
// *
