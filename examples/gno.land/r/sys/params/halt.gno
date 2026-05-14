package params

import (
	"strconv"

	"chain"
	prms "sys/params"

	"gno.land/r/gov/dao"
)

const (
	nodeModulePrefix  = "node"
	haltHeightKey     = "halt_height"
	haltMinVersionKey = "halt_min_version"
)

// NewSetHaltRequest creates a GovDAO proposal to halt all chain nodes at the given block height.
// Once approved and executed, nodes will gracefully stop after committing the specified block,
// enabling coordinated chain upgrades.
//
// minVersion, if non-empty, sets the minimum binary version required to resume after the halt.
// Nodes will refuse to restart unless their version satisfies the minimum requirement,
// preventing old binaries from accidentally resuming a chain halted for an upgrade.
// Example: minVersion="chain/gnoland1.1" prevents gnoland1.0 from resuming.
//
// Use height=0 to cancel a previously scheduled halt.
func NewSetHaltRequest(height int64, minVersion string) dao.ProposalRequest {
	callback := func(cur realm) error {
		prms.SetSysParamInt64(nodeModulePrefix, "p", haltHeightKey, height)
		prms.SetSysParamString(nodeModulePrefix, "p", haltMinVersionKey, minVersion)
		chain.Emit("set_halt",
			"height", strconv.FormatInt(height, 10),
			"min_version", minVersion,
		)
		return nil
	}

	var desc string
	if height == 0 {
		desc = "Cancel the scheduled chain halt and clear the minimum version requirement."
	} else {
		desc = "Halt the chain at block " + strconv.FormatInt(height, 10) + "."
		if minVersion != "" {
			desc += " Requires binary version >= " + minVersion + " to resume."
		}
	}

	e := dao.NewSimpleExecutor(callback, "")
	return dao.NewProposalRequest("Set node halt height", desc, e)
}
