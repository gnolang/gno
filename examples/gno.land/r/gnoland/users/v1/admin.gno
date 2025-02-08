package v1

import (
	"gno.land/p/demo/dao"

	"gno.land/r/gov/dao/bridge"
)

var paused = false // until a DAO can be owner of ownable/pausable

// NewSetPausedExecutor allows GovDAO to pause or unpause this realm
func NewSetPausedExecutor(newPausedValue bool) dao.Executor {
	cb := func() error {
		paused = newPausedValue
		return nil
	}

	return bridge.GovDAO().NewGovDAOExecutor(cb)
}
