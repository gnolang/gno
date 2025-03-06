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
	trees     = avl.NewTree() // string(root ID) -> avl.Tree(string(path) -> *commondao.CommonDAO)
)

// i get the public.gno file, but imho its outdated per r/demo/boards
// let's try leaving the main logic of the package in the same file name as the package, ie commondao.gno
// when i look into the commondao.gno i expect to see the main functions of the package
// as a user i still want to understand whats going on, and not look at some public.gno to 
// s ee only the functions that i can call

func getDAO(daoID uint64) *commondao.CommonDAO {
	key := makeIDKey(daoID)
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

func getSubDAO(rootID uint64, path string) *commondao.CommonDAO {
	tree := getTree(rootID)
	if tree == nil {
		return nil
	}

	if v, found := tree.Get(path); found {
		return v.(*commondao.CommonDAO)
	}
	return nil
}

func mustGetSubDAO(rootID uint64, path string) *commondao.CommonDAO {
	dao := getSubDAO(rootID, path)
	if dao == nil {
		panic("subDAO not found")
	}
	return dao
}

func getTree(rootID uint64) *avl.Tree {
	key := makeIDKey(rootID)
	if v, found := trees.Get(key); found {
		return v.(*avl.Tree)
	}
	return nil
}

func getOwnership(addr std.Address) []uint64 {
	if v, ok := ownership.Get(addr.String()); ok {
		return v.([]uint64)
	}
	return nil
}

func makeIDKey(daoID uint64) string {
	return seqid.ID(daoID).String()
}
