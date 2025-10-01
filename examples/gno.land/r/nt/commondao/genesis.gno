package commondao

import (
	"chain/runtime"

	"gno.land/p/nt/commondao"
)

func init() {
	// Create a DAO for the realm
	id := daoID.Next()
	dao := commondao.New(
		commondao.WithID(uint64(id)),
		commondao.WithName("Common DAO"),
		commondao.WithDescription(
			"This DAO is responsible for managing `commondao` realm functionalities.",
		),
	)

	// Add initial members
	dao.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	dao.Members().Add("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul

	// Index DAO
	daos.Set(id.String(), dao)

	// Make commondao realm the owner of Common DAO
	commonRealm := runtime.CurrentRealm()
	ownership.Set(commonRealm.Address().String(), []uint64{CommonDAOID})

	// Save DAO configuration
	options.Set(id.String(), &Options{
		AllowListing:       true,
		AllowRender:        true,
		AllowVoting:        true,
		AllowExecution:     true,
		AllowMembersUpdate: true,
	})
}
