package params

import (
	prms "sys/params"

	"gno.land/r/gov/dao"
)

const (
	nodeModulePrefix = "node"
	haltHeightKey    = "halt_height"
)

// NewSetHaltHeightRequest creates a GovDAO proposal to halt all chain nodes at the given block height.
// Once the proposal is approved and executed, all nodes reading the halt_height parameter will
// gracefully stop after committing the specified block, enabling coordinated chain upgrades.
//
// Use height=0 to cancel a previously scheduled halt.
func NewSetHaltHeightRequest(height int64) dao.ProposalRequest {
	key := syskey(nodeModulePrefix, "p", haltHeightKey)
	return newPropRequest(
		key,
		func() { prms.SetSysParamInt64(nodeModulePrefix, "p", haltHeightKey, height) },
		"Set node halt height",
	)
}
