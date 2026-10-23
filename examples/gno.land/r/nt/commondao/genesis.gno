package commondao

func init() {
	// Fail closed if deployed off the package path treasury addresses
	// derive from (see assertRunningPath).
	assertRunningPath()

	// Create the realm's own DAO, governed by its maintainers. There is no
	// owner: the realm administers nothing about a DAO except its home-index
	// listing, which the realm DAO sets for itself here.
	dao := createDAO(
		"Common DAO",
		"Govern and maintain the commondao reference realm.",
		"This DAO is responsible for managing `commondao` realm functionalities.",
		"g1skl80cuz8zq3lul9pgz5pc35l2pfzgxgfpsqkx", // govdao t1 multisig
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)
	setListed(dao.ID(), true)
}
