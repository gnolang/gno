package govdao

import "gno.land/p/system/govtypes"

type GovDAO struct {
	// tiered members
}

// ensure GovDAO implements the types.Node interface.
var _ = types.Node(DAO{})

/*
- gov/directory        -> explore DAOs and Actors in the network
- gov/govdao           -> top-level DAO node with tiered members
- gov/types,interfaces ->
- worx/profiles        -> portfolios built around contributions, endorsements, fact oracles and personal presentations
- worx/evaluation      -> contributions management
- worx/dao

- dxteam/dao
- berty/dao
- teritori/dao
- onbloc/dao
*/
