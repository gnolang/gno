package govdao

import "std"

// DIRECTORY
var head Project

// GOVDAO
type GovDAO struct{}
type Tier struct{}

// GOVTYPES

type Project interface {
	Children() []Project
	Members() []Actor
}

// Actor XXX
// Actor can represent a blockchain account (std.Address), a realm (PkgPath), or a go object.
type Actor interface{}

// an object COULD be Project AND Actor

func NewTeamAnyActor() Actor                 {}
func NewPackageActor(pkgpath string) Actor   {}
func NewAccountActor(addr std.Address) Actor {}
func NewGoActor(impl interface{}) Actor      {} // it is advised to create its own implementation matching a specific interface

// Account (std.Address)  // human, multisig, oracles
// Realm (std.PkgPath)    // realm
// GoObject (interface{}) // a dao, whatever

// workflow -> anyone can create something matching the interface
// it's up to the

/*
   - gov/directory        -> explore DAOs and Actors in the network
   - gov/govdao           -> top-level DAO node with tiered members
   - gov/types,interfaces ->
   - worx/profiles        -> portfolios built around contributions, endorsements, fact oracles and personal presentations
   - worx/evaluation      -> contributions management
   - worx/dao

   - dxteam/dao
   - berty/dao
   - teritori/dao
   - onbloc/dao
*/

var directory Node

// govdao
//   worxdao -> autonomous but also mandated by govdao; HR; "paying salaries"
//   anotherchain

func init() {
	var govdao Node
	var personJae Node
	var personManfred Node
	var worxDAO Node
	var companyNT Node // distribute treasury to employees
	var chainCosmos Node
	var projTM2 Node
	var projTM2Amino Node
	var projGnovm Node
	var projGnoland Node

	// govdao.AttachChild(worxDAO)
	// govdao.Proposal(addNewChild)
}

type Contribution struct {
	Links []string
}

type Contributor struct {
	Tier  int
	Score int
}

// worxdao -> HR => bonus, hire, investment, fire, promote, demote,
// earn points/money to review

// worxDAO -> bucket
// levels are managing levels below; they are filtering, reviewing, and score
// proposal ->

// tokens ^worx
// tokens $gnot
// tokens ^gov

type Directory interface {
	Node // embeds a Node
	// ...
}

type Node interface {
	// Owner Node
	// Facilitator Node
}

func main() {
	println("TEST2")
}

type EvaluationDAO struct{}

type GovDAO struct{}

type Manfred struct{}

type NT struct{}

// XXX: switch to Book
// split
// multiple governance; one of them is following levels, not all of them

// 70% contributors
// 20% validators
// 5% contracts
// 5% holders / freezer

/*var (
	projects []Project
	teams    []Team
	people   []Person
        )
*/

/*
type Project interface {
	owner   Team
	members Team
}

type Team struct {
	people []Person
}

type Person struct {
	addr std.Address
}
*/

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
