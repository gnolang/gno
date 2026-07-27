package commondao

import (
	"chain"
	"strconv"

	"gno.land/p/moul/txlink"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/seqid/v0"
)

// pkgPath is this realm's package path, fixed regardless of quarantine
// location. DAO treasury addresses derive from it, so they are stable
// across logic upgrades.
const pkgPath = "gno.land/r/nt/commondao/v0"

// CommonDAOID is the ID of the realm's DAO.
const CommonDAOID uint64 = 1

var (
	daoID     seqid.ID
	realmLink = txlink.Realm(pkgPath)
	daos      = bptree.NewBPTree32() // string(ID) -> *commondao.CommonDAO
	ownership = bptree.NewBPTree32() // string(address) -> []uint64(DAO ID)
	invites   = bptree.NewBPTree32() // string(address) -> address(inviter)
	options   = bptree.NewBPTree32() // string(ID) -> *Options
)

func getDAO(daoID uint64) *commondao.CommonDAO {
	key := makeIDKey(daoID)
	if v := daos.Get(key); v != nil {
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

func getOwnership(addr address) []uint64 {
	if v := ownership.Get(addr.String()); v != nil {
		return v.([]uint64)
	}
	return nil
}

func getOptions(daoID uint64) *Options {
	key := makeIDKey(daoID)
	if v := options.Get(key); v != nil {
		return v.(*Options)
	}
	return nil
}

func makeIDKey(daoID uint64) string {
	return seqid.ID(daoID).String()
}

// subpathOf returns the realm sub-identity subpath that acts for a DAO.
func subpathOf(daoID uint64) string {
	return "dao/" + strconv.FormatUint(daoID, 10)
}

// daoAddress derives a DAO's treasury address: the address of the realm
// sub-identity cur.Sub(subpathOf(daoID)) that executors mint to move its
// funds. Pure derivation; DAO IDs are never reused, so the address is
// collision free and stable.
func daoAddress(daoID uint64) address {
	return chain.DerivePkgSubAddr(pkgPath, subpathOf(daoID))
}

func createDAO(name, description string, owner address, members ...address) *commondao.CommonDAO {
	if len(members) == 0 {
		panic("a DAO requires at least one initial council member")
	}

	id := daoID.Next()
	dao := commondao.New(newDAOOptions(uint64(id), name, description, nil, members)...)

	ids := append(getOwnership(owner), uint64(id))
	ownership.Set(owner.String(), ids)

	daos.Set(id.String(), dao)

	stored := defaultOptions
	options.Set(id.String(), &stored)
	return dao
}

func createSubDAO(parent *commondao.CommonDAO, name, description string, members ...address) *commondao.CommonDAO {
	if len(members) == 0 {
		panic("a SubDAO requires at least one initial council member")
	}

	// WithParent also registers the new DAO as one of parent's children
	id := daoID.Next()
	dao := commondao.New(newDAOOptions(uint64(id), name, description, parent, members)...)

	stored := defaultOptions
	options.Set(id.String(), &stored)
	daos.Set(id.String(), dao)
	return dao
}

func newDAOOptions(id uint64, name, description string, parent *commondao.CommonDAO, members []address) []commondao.Option {
	opts := []commondao.Option{
		commondao.WithID(id),
		commondao.WithName(name),
		commondao.WithDescription(description),
		commondao.WithAddress(daoAddress(id)),
	}
	if parent != nil {
		opts = append(opts, commondao.WithParent(parent))
	}
	for _, m := range members {
		opts = append(opts, commondao.WithCouncilMember(m))
	}
	return opts
}
