package bridge

import (
	"std"

	"gno.land/p/demo/ownable"
)

const initialOwner = std.Address("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @moul

var b *Bridge

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
		dao:     &govdaoV2{},
	}
}

// SetDAO sets the currently active GovDAO implementation
func SetDAO(dao DAO) {
	b.AssertCallerIsOwner()

	b.dao = dao
}

// GovDAO returns the current GovDAO implementation
func GovDAO() DAO {
	return b.dao
}
