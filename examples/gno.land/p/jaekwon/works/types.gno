package works

import "std"

//----------------------------------------
// Objective

// TODO rename
type ObjectiveID string 

type Objective struct {
	ObjectiveID
	CreatedAt time.Time
	Components avl.Tree
}

//----------------------------------------
// Treasury

type Treasury interface {
	Balance() std.Coins
	Bylaws() Bylaws
	// XXX what else goes here?
}

// Standard charter structure
type Bylaws struct {
	// XXX
}

//----------------------------------------
// Contract
// Represents a contract with objectives 

type Contract struct {
	CreatedAt time.Time
	Objectives []*Objective
	// XXX
}

//----------------------------------------
// WorkGroup

type WorkGroupID string

type WorkGroup struct {
	WorkGroupID
	CreatedAt time.Time
	Objectives []*Objective
	Treasury
	Bylaws
	// Forums
	// Data
}
