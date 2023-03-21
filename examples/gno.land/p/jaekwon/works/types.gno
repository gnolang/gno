package works

import "std"

//----------------------------------------
// Objective

// TODO rename
type ObjectiveID string 

type Objective struct {
	id ObjectiveID
	createdAt time.Time
	components avl.Tree
	parent *Objective
	description string 
	contracts avl.Tree // ContractID -> Contract ??? competing contracts implementing objective.
	related avl.Tree // string ID -> interface{} ??? all other related objects.
}

// TODO Similar to Treasury, objectives spawn new components?

//----------------------------------------
// Treasury

type TreasuryID ObjectID

type (tid TreasuryID) String() string {
	return "t:"+ObjectID(tid).String()
}

type Treasury struct {
	id TreasuryID
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

type ContractID ObjectID

type (cid ContractID) String() string {
	return "c:"+ObjectID(cid).String()
}

type Contract struct {
	id ContractID
	createdAt time.Time
	objectives []*Objective // XXX can refer to external objectives?
	// XXX
	// signers?
	// signed?
}

// TODO master contract, ongoing work updates etc?

//----------------------------------------
// Workgroup

type WorkgroupID ObjectID

type (wid WorkgroupID) String() string {
	return "w:"+ObjectID(wid).String()
}

type Workgroup struct {
	id WorkgroupID
	createdAt time.Time
	objectives []*Objective
	treasury Treasury
	bylaws Bylaws
	// forum Forum
	// data Data
}

type Bylaws struct {
	// XXX
}

