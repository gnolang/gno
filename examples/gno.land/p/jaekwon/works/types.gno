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
	Description string 
}

//----------------------------------------
// Treasury

type Treasury struct {
	total std.Coins // total for all accounts including all accounts.
	balance std.Coins // account for this treasury. 0 <= balance <= total.
	accounts avl.Tree // initially empty.
}

func (tr *Treasury) NewAccount(...) *Treasury {
	// TODO create new treasury and associate with tr.accounts.
	panic("not implemented")
	return nil
}

func (tr *Treasury) VerifyIntegrity() {
	// TODO go through accounts and ensure that all internal accounts
	// are nonnegative and sum to tr.total.
	// TODO supplant/replace with "avl.ReduceTree" to allow reduction directly in avl.
	//      that would be more scalable.
	panic("not implemented")
}

//----------------------------------------
// Contract
// Represents a contract with objectives 

type Contract struct {
	CreatedAt time.Time
	Objectives []*Objective // XXX can refer to external objectives?
	// XXX
	// signers?
	// signed?
}

//----------------------------------------
// Workgroup

type WorkgroupID string

type Workgroup struct {
	WorkgroupID
	CreatedAt time.Time
	Objectives []*Objective
	Treasury
	Bylaws
	// Forums
	// Data
}

type Bylaws struct {
	// XXX
}

