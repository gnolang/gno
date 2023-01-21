package works

import "std"

// TODO: String ID, StringID() string is maybe inferior to
// ensuring that ObjectiveID etc are a priori unique via hashing.

//----------------------------------------
// Objective

// TODO rename
type ObjectiveID string 

type Objective struct {
	ObjectiveID
	CreatedAt time.Time
	Components avl.Tree
	Parent *Objective
	Description string 
	Contracts avl.Tree // ContractID -> Contract ??? competing contracts implementing objective.
	Related avl.Tree // string ID -> interface{} ??? all other related objects.
}

// TODO Similar to Treasury, objectives spawn new components?

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

// XXX Treasury is an Account interface?
// XXX An Account is something you can deposit into?
// XXX Or Withdraw from for purposes?
// XXX Treasury vs ?

func (tr *Treasury) VerifyTotal() error {
	// TODO go through accounts and ensure that all internal accounts
	// are nonnegative and sum to tr.total.
	// TODO supplant/replace with "avl.ReduceTree" to allow reduction directly in avl.
	//      that would be more scalable.
	panic("not implemented")
}

//----------------------------------------
// Contract
// Represents a contract with objectives 

type ContractID string

type (cid ContractID) StringID() string {
	return "contract:"+cid
}

type Contract struct {
	CreatedAt time.Time
	Objectives []*Objective // XXX can refer to external objectives?
	// XXX
	// signers?
	// signed?
}

// TODO master contract, ongoing work updates etc?

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

