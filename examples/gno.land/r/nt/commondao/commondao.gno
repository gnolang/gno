package commondao

import (
	"chain/runtime"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/commondao"
	"gno.land/p/nt/seqid"
)

// CommonDAOID is the ID of the realm's DAO.
const CommonDAOID uint64 = 1

var (
	daoID     seqid.ID
	realmLink txlink.Realm
	daos      = avl.NewTree() // string(ID) -> *commondao.CommonDAO
	ownership = avl.NewTree() // string(address) -> []uint64(DAO ID)
	invites   = avl.NewTree() // string(address) -> address(inviter)
	trees     = avl.NewTree() // string(root ID) -> avl.Tree(string(DAO ID) -> *commondao.CommonDAO)
	options   = avl.NewTree() // string(ID) -> *Options
)

func init() {
	// Save current realm path so it's available during render calls
	realmLink = txlink.Realm(runtime.CurrentRealm().PkgPath())
}

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

func getTree(rootID uint64) *avl.Tree {
	key := makeIDKey(rootID)
	if v, found := trees.Get(key); found {
		return v.(*avl.Tree)
	}
	return nil
}

func getOwnership(addr address) []uint64 {
	if v, ok := ownership.Get(addr.String()); ok {
		return v.([]uint64)
	}
	return nil
}

func getOptions(daoID uint64) *Options {
	key := makeIDKey(daoID)
	if v, found := options.Get(key); found {
		return v.(*Options)
	}
	return nil
}

func makeIDKey(daoID uint64) string {
	return seqid.ID(daoID).String()
}

func createDAO(name string, owner address, o []Option) *commondao.CommonDAO {
	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithName(name),
	)

	daoOptions := defaultOptions
	for _, apply := range o {
		apply(&daoOptions)
	}

	ids := append(getOwnership(owner), uint64(id))
	ownership.Set(owner.String(), ids)

	daos.Set(id.String(), dao)
	options.Set(id.String(), &daoOptions)
	return dao
}

func createSubDAO(parent *commondao.CommonDAO, name string, o []Option) *commondao.CommonDAO {
	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithParent(parent),
		commondao.WithName(name),
	)

	daoOptions := defaultOptions
	for _, apply := range o {
		apply(&daoOptions)
	}

	options.Set(id.String(), &daoOptions)
	daos.Set(id.String(), dao)

	rootID := parent.TopParent().ID()
	tree := getTree(rootID)
	if tree == nil {
		tree = avl.NewTree()
		trees.Set(makeIDKey(rootID), tree)
	}

	tree.Set(id.String(), dao)
	return dao
}
