package bridge

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/ownable"
)

const initialOwner = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @moul

var (
	b                 *Bridge
	allowedGovDAOPath = "gno.land/r/gov/dao/v2" // set as initial value

	ErrInvalidGovDAOPkgPath = errors.New("invalid GovDAO package path")
)

// Bridge is the active GovDAO
// implementation bridge
type Bridge struct {
	*ownable.Ownable

	dao DAO
}

// init constructs the initial GovDAO implementation
func init() {
	b = &Bridge{
		Ownable: ownable.NewWithAddress(initialOwner),
		dao:     nil, // set upon r/gov/dao/v2 init()
	}
}

// AllowNewGovDAOVersion allowlists the new GovDAO version
// This allows the GovDAO contract to call RegisterNewDAOImpl inside init()
func AllowNewGovDAOVersion(newGovDAOPath string) {
	b.AssertCallerIsOwner()

	if !strings.HasPrefix(newGovDAOPath, "gno.land/r/gov/dao/") {
		panic(ErrInvalidGovDAOPkgPath)
	}

	allowedGovDAOPath = newGovDAOPath
}

// RegisterNewDAOImpl allows a GovDAO realm to register itself to the bridge
// after it has been allowlisted by AllowNewGovDAOVersion
func RegisterNewDAOImpl(d DAO) {
	if std.PrevRealm().PkgPath() != allowedGovDAOPath {
		panic("caller not allowlisted")
	}

	b.dao = d
}

// GovDAO returns the current GovDAO implementation
func GovDAO() DAO {
	return b.dao
}
