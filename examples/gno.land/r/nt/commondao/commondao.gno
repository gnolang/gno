package commondao

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
	"gno.land/p/nt/commondao"
)

// TODO: Use a trie tree to store hierarchical DAOs

var (
	daoID     seqid.ID
	daos      = avl.NewTree() // string(ID) -> *commondao.CommonDAO
	ownership = avl.NewTree() // string(std.Address) -> []uint64(DAO ID)
	invites   = avl.NewTree() // string(std.Address) -> std.Address(inviter)
	trees     = avl.NewTree() // string(root ID + ":" + path) -> *commondao.CommonDAO
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

func getOwnership(addr std.Address) []uint64 {
	if v, ok := ownership.Get(addr.String()); ok {
		return v.([]uint64)
	}
	return nil
}

func makeTreesKey(rootID uint64, path string) string {
	return seqid.ID(rootID).String() + ":" + path
}
