package multi_sig

import (
	"std"
)

type Function interface{}
type Parameters interface{}

type FunctionProposal struct {
	id        int
	function  Function
	params    Parameters
	approvals int
}

var (
	owners    []std.Address
	proposals []FunctionProposal
)

func AddProposal(f Function, p Parameters) {

}

func isOwner(caller std.Address) bool {
	for _, owner := range owners {
		if caller == owner {
			return true
		}
	}
	return false
}