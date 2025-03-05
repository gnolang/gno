package commondao

import (
	"std"

	"gno.land/p/nt/commondao"
)

func init() {
	// Create a DAO for the realm
	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithName("Common DAO"),
	)

	// Add initial members
	dao.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	dao.Members().Add("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul

	// Index DAO
	daos.Set(id.String(), dao)

	// Make commondao realm the owner of Common DAO
	realm := std.CurrentRealm()
	ownership.Set(realm.Address().String(), []uint64{CommonDAOID})
}
