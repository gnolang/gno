package govdao

import (
	"std"
	"strconv"
)

func main() {
	println(govDAO.Directory(""))
}

var govDAO = GovDAO{}

func init() {
	var (
		treasuryDAO      = BaseDAO{name: "TreasuryDAO", parent: govDAO}
		worxDAO          = BaseDAO{name: "WorxDAO", parent: govDAO}
		evaluationDAO    = BaseDAO{name: "EvaluationDAO", parent: worxDAO, comment: ""}
		engineeringDAO   = BaseDAO{name: "EngineeringDAO", parent: worxDAO, comment: "responsible for high-level architecture"}
		grantsDAO        = BaseDAO{name: "grantsDAO", parent: worxDAO, comment: "manages grants"}
		projectsDAO      = BaseDAO{name: "ProjectsDAO", parent: worxDAO, comment: "reverse of a grant: we lock money for a cause, then it can be used by a selected committee"}
		marketingDAO     = BaseDAO{name: "marketingDAO"}
		savethecosmosDAO = BaseDAO{name: "savethecosmosDAO"}
	)

	govDAO.tiers = []*Tier{{idx: 0}, {idx: 1}, {idx: 2}}
	govDAO.tiers[0].parent = nil
	govDAO.tiers[0].child = govDAO.tiers[1]
	govDAO.tiers[1].parent = govDAO.tiers[0]
	govDAO.tiers[1].child = govDAO.tiers[2]
	govDAO.tiers[2].parent = govDAO.tiers[1]
	govDAO.tiers[2].child = nil
	govDAO.children = []DAO{treasuryDAO, worxDAO}
	worxDAO.children = []DAO{evaluationDAO, engineeringDAO, projectsDAO}
	projectsDAO.children = []DAO{savethecosmosDAO, grantsDAO}
	engineeringDAO.children = []DAO{
		BaseDAO{
			name: "tm2", parent: worxDAO, children: []DAO{
				BaseDAO{name: "amino"},
				BaseDAO{name: "bft"},
			},
		},
		BaseDAO{
			name: "gnovm", parent: worxDAO, children: []DAO{
				BaseDAO{name: "gnolang"},
				BaseDAO{name: "stdlibs"},
			},
		},
		BaseDAO{name: "gno.land", parent: worxDAO},
	}
	grantsDAO.children = []DAO{
		BaseDAO{name: "onbloc"},
		BaseDAO{name: "berty"},
		BaseDAO{name: "teritori"},
		BaseDAO{name: "grant1"},
		BaseDAO{name: "grant2"},
	}
}

type GovDAO struct {
	tiers    []*Tier
	children []DAO
}

type DAO interface {
	Directory(prefix string) string
	Parent() DAO
	HasChild(DAO) bool
}

func (d GovDAO) Parent() DAO { return nil }
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

// func (t *Tier) Promote()
// func (t *Tier) Demote()
func (t Tier) Directory(prefix string) string {
	output := prefix + "- tier#" + strconv.Itoa(t.idx) + "\n"
	return output
}

type BaseDAO struct {
	name     string
	children []DAO
	parent   DAO
	comment  string
}

func (d BaseDAO) Directory(prefix string) string {
	output := prefix + "- " + d.name
	if d.comment != "" {
		output += " // " + d.comment
	}
	output += "\n"
	childPrefix := prefix + "  "

	for _, child := range d.children {
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
