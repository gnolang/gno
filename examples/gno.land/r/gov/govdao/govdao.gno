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
