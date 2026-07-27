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
	invites   = bptree.NewBPTree32() // string(address) -> address(inviter)
	creators  = bptree.NewBPTree32() // string(address) -> struct{}; consumed-invite marker
	listed    = bptree.NewBPTree32() // string(ID) -> struct{}; home-index opt-in
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

func makeIDKey(daoID uint64) string {
	return seqid.ID(daoID).String()
}

// isCreator reports whether an address has already consumed an invite and
// may create further DAOs without another one.
func isCreator(addr address) bool {
	return creators.Has(addr.String())
}

// isListed reports whether a DAO opted into the realm's public home index.
func isListed(daoID uint64) bool {
	return listed.Has(makeIDKey(daoID))
}

// setListed adds or removes a DAO from the realm's public home index.
func setListed(daoID uint64, v bool) {
	key := makeIDKey(daoID)
	if v {
		listed.Set(key, struct{}{})
	} else {
		listed.Remove(key)
	}
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

func createDAO(name, description string, members ...address) *commondao.CommonDAO {
	if len(members) == 0 {
		panic("a DAO requires at least one initial council member")
	}

	id := daoID.Next()
	dao := commondao.New(newDAOOptions(uint64(id), name, description, nil, members)...)
	daos.Set(id.String(), dao)
	return dao
}

func createSubDAO(parent *commondao.CommonDAO, name, description string, members ...address) *commondao.CommonDAO {
	if len(members) == 0 {
		panic("a SubDAO requires at least one initial council member")
	}

	// WithParent also registers the new DAO as one of parent's children
	id := daoID.Next()
	dao := commondao.New(newDAOOptions(uint64(id), name, description, parent, members)...)
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
