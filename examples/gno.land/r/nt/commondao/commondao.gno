package commondao

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/nt/commondao"
)

var (
	daoID   seqid.ID
	daos    = avl.NewTree() // string(DAO ID) -> *commondao.CommonDAO
	invites = avl.NewTree() // string(std.Address) -> std.Address(inviter)
)

func getDAO(daoID uint64) *commondao.CommonDAO {
	key := seqid.ID(daoID).String()
	if v, found := daos.Get(key); found {
		return v.(*commondao.CommonDAO)
	}
	return nil
}

func mustGetDAO(daoID uint64) *commondao.CommonDAO {
	dao := getDAO(daoID)
	if dao == nil {
		panic("DAO not found")
	}
	return dao
}
