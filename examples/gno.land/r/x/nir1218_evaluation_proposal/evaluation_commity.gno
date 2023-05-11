package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

/*
Contributors DAO will designate its memebers to an Evaluation Commity, in the begining, the evaluation commity members could be the core development team members or any other trusted entity.
Evaluation Commity sets the contribution evaluation criteria for a pull request, based on a pull request type or scope, and its corresponding contribution.
Once the contribution evaluaiton criteria is set, each member of the evaluation commity can evaluate a contribution by giving a certain number of points.
Scoring a contribution is essential when there are competing contributions (Game of Realm). Otherwise, the evaluaiton is a binary decision.
Based on the total score, the evaluation commity can negotiate and later approve a contribution.
*/

/*
 Evaluation Commity is a list of addresses of its memebers
 Evaluation Commity sets contribution types
 Evaluation Commity sets evalution criteria per pull request type and scope
 Evaluation Commity sets evalution criteria scoring range (1-10)
 Evaluation Commity negotiates and approves contributions
*/

type EvaluationCommity struct {
	members []std.Address
}

func NewEvalutaionCommity() *EvaluationCommity {
	ec := &EvaluationCommity{
		members: []std.Address{},
	}
	return ec
}

func (ec *EvaluationCommity) designateMembers(members []std.Address) []std.Address {
	ec.members = append(ec.members, members...)
	return ec.members
}

func (ec *EvaluationCommity) dismissMembers(members []std.Address) []std.Address {
	// TODO
}

func equal(a, b []std.Address) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
