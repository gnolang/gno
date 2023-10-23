package govtypes

import "std"

// type TreeNode // wrapper that adds the Parent() by recursively checking the below interface

// Project XXX
type Project interface {
	Parent() Project
	Children() Project
}

// Actor XXX
// Actor can represent a blockchain account (std.Address), a realm (PkgPath), or a go object.
type Actor interface {
	// XXX embeds a p/ something with account abstraction
	ID()
}

func NewPackageActor(pkgpath string) Actor   {}
func NewAccountActor(addr std.Address) Actor {}
func NewGoActor(impl interface{}) Actor      {} // it is advised to create its own implementation matching a specific interface

// Account (std.Address)  // human, multisig, oracles
// Realm (std.PkgPath)    // realm
// GoObject (interface{}) // a dao, whatever

// workflow -> anyone can create something matching the interface
// it's up to the
