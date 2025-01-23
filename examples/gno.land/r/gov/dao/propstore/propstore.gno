package propstore

import (
	"std"

	"gno.land/r/gov/dao/ruler"
)
var proposals []*Proposal

var callback DAOCallback

func init(){
	// TODO: set current DAOCallback impl
}

// MustNewProposal is like NewProposal but it panics if the proposal cannot be added.
func MustNewProposal(title string, desc string, callback func(Metadata) error, meta Metadata) int {
	id ,err := NewProposal(title, desc, callback, meta)
	if err != nil {
		panic(err)
	}

	return id
}

// NewProposal creates a new proposal with the specified information.
// The provided callback will be executed when the proposal is validated by GovDAO.
// The actual govDAO implementation will be in charge of validating that the proposer 
// is able to create the proposal or not. Metadata object is an arbitrary object that 
// will be sent to GovDAO for validation purposes and also it will be sent to the callback function.
// That metadata object can be used to give extra information to the callback function if needed, or even give 
// some metadata information to the actual GovDAO in charge for validation purposes.
func NewProposal(title string, desc string, callback func(Metadata) error, meta Metadata) (int, error) {
	caller := std.GetOrigCaller()

	if err := callback.BeforeNewProposal(caller,title, meta); err != nil {
		return -1, err
	}
	
	p := &Proposal{
		Author: caller, 
		Title: title,
		Description: desc,
		Executor: newExecutor(callback, meta),
		RulerOnCreation: ruler.CurrentDAO,
	}

	pid = len(proposals)
	proposals = append(proposals, p)
	callback.AfterNewProposal(caller,title, meta, pid)

	return pid, nil
}

// GetProposal gets a proposal by ID
func GetProposal(id int) *Proposal {
	if id > len(proposals)-1 {
		return nil
	}

	return proposals[id]
}

// SetCallback gives to the actual GovDAO in charge the permissions 
// to set a DAOCallback implementation, giving information about when 
// a Proposal is created, giving the opportunity to GovDAO to validate 
// Proposal creations.
func SetCallback(cb DAOCallback) {
	if !checkCurrent(){
		panic("this method cann only be called by current GovDAO")
	}

	callback = cb
}

func caller() string {
	return std.PrevRealm().PkgPath()
}

func checkCurrent() bool {
	return caller() == ruler.CurrentDAO
}