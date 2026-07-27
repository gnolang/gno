package commondao

func init(cur realm) {
	// Create the realm's own DAO. cur is the chain-injected init cur:
	// cur.Address() is the commondao realm's package address, so the DAO
	// is owned by the realm itself and its options are frozen after init.
	createDAO(
		"Common DAO",
		"This DAO is responsible for managing `commondao` realm functionalities.",
		cur.Address(),
		"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq", // @devx
		"g1manfred47kzduec920z88wfr64ylksmdcedlf5", // @moul
	)

	// The realm DAO is listed but does not host sub-DAOs
	o := getOptions(CommonDAOID)
	o.AllowListing = true
	o.AllowChildren = false
}
