package v1

var paused = false // until a DAO can be owner of ownable/pausable

// FIXME: Blocked by https://github.com/gnolang/gno/issues/3546
//// NewPauseUnpauseExecutor allows GovDAO to pause or unpause this realm
//func NewPauseUnpauseExecutor(newPausedValue bool) dao.Executor {
//	cb := func() error {
//		paused = newPausedValue
//		return nil
//	}
//
//	return bridge.GovDAO().NewGovDAOExecutor(cb)
//}
