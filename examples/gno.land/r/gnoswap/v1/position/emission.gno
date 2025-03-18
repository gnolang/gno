package position

import (
	br "gno.land/r/gnoswap/v1/bridge"
)

// mintAndDistributeGnsCallback executes the MintAndDistributeGns callback
//
// NOTE: Currently, this function is a demo version that applies the callback pattern
// to avoid import-path dependencies.
//
// Previously, when upgrading a contract, there was an issue where it couldn't reference new contracts
// because the package paths of imported contracts were fixed.
// Therefore, we implemented this pattern of registering functions as callbacks through a bridge and then calling them.
//
// If security and functionality issues are confirmed to be absent,
// it will later be modified to directly call callbacks instead of using separate wrapper functions as it does now,
// and other emission-dependent functions will also be handled with callbacks like this case.
func mintAndDistributeGnsCallback() uint64 {
	cb, exists := br.GetCallback(br.MINT_AND_DISTRIBUTE_GNS_KEY)
	if !exists {
		panic("MintAndDistributeGns callback is not set!")
	}

	mintCallback, ok := cb.(func() uint64)
	if !ok {
		panic("Invalid callback type for MintAndDistributeGns")
	}

	return mintCallback()
}
