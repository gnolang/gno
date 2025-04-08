package home

import (
	"gno.land/p/moonia/dao"
)

var ds DAOState

type DAOState struct {
	DAO       *dao.DAO
	Proposals *dao.ProposalStore
}

func init() {
	myDao := dao.NewDAO()
	myProps := dao.NewProposalStore(myDao)

	ds = DAOState{
		DAO:       myDao,
		Proposals: myProps,
	}
}