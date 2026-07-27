package commondao

func init(cur realm) {
	// Create the realm's own DAO, governed by its maintainers. There is
	// no owner: after the ownership re-scope the realm administers nothing
	// about a DAO except its home-index listing, which the realm DAO sets
	// for itself here.
	dao := createDAO(
		"Common DAO",
		"This DAO is responsible for managing `commondao` realm functionalities.",
		"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", // @devx
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)
	setListed(dao.ID(), true)
}
