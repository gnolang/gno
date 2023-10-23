package dao

import "std"

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
