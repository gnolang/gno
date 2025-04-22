package dao

import "fmt"

const (
	DAONameLead        = "lead"
	DAONameSupervisory = "super"
)

func DAOName(name string, i int) string {
	if i == 0 {
		return name
	} else {
		return fmt.Sprintf("%s#%d", name, i)
	}
}

type defaultDAO struct {
	subDAOs avl.Tree
	bylaws  string
	passed  avl.Tree
}

func (dao *defaultDAO) GetSubDAO(name string) DAO {
	// ...
}

//----------------------------------------

type Proposal struct {
	status    Status
	text      string
	approvals []approval // XXX
	comments  []comment  // XXX tree?
}

//----------------------------------------
/*



 */

type TopDAO struct {
	defaultDAO
	params *Params

	gov      // gov XXX
	steering // steering

	oversight  // oversight
	validators // validators XXX
}

type ProposalActivator func(*Proposal) error

func (*TopDAO) ParamQualFn() ProposalActivator {
	return func(prop *Proposal) error {
		// must pass oversight.
		// must pass proposal.
		// must pass quorum.
		// change status
	}
}

//----------------------------------------

type DAO interface {
	// ...
}

func main() {
}
