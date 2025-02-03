package bridge

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ownable"
)

const (
	initialOwner = std.Address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
	loader       = "gno.land/r/gov/dao/init"
)

var (
	b       *Bridge
	Ownable = ownable.NewWithAddress(initialOwner)
)

// Bridge is the active GovDAO
// implementation bridge
type Bridge struct {
	dao DAO
}

// init constructs the initial GovDAO implementation
func init() {
	b = &Bridge{
		dao: nil, // initially set via r/gov/dao/loader
	}
}

// LoadGovDAO loads the initial version of GovDAO into the bridge
// All changes to b.dao need to be done via GovDAO proposals after
func LoadGovDAO(d DAO) {
	if std.PrevRealm().PkgPath() != loader {
		panic("unauthorized")
	}

	b.dao = d
}

// NewGovDAOImplChangeExecutor allows creating a GovDAO proposal
// Which will upgrade the GovDAO version inside the bridge
func NewGovDAOImplChangeExecutor(newImpl DAO) dao.Executor {
	callback := func() error {
		b.dao = newImpl
		return nil
	}

	return b.dao.NewGovDAOExecutor(callback)
}

// SetGovDAO allows the admin to set the GovDAO version manually
// This functionality can be fully disabled by Ownable.DropOwnership(),
// making this realm fully managed by GovDAO.
func SetGovDAO(d DAO) {
	Ownable.AssertCallerIsOwner()
	b.dao = d
}

// GovDAO returns the current GovDAO implementation
func GovDAO() DAO {
	return b.dao
}
