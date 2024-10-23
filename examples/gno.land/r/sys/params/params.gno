package params

import (
	"std"

	"gno.land/p/gov/proposal"
)

const daoPkgPath = "gno.land/r/gov/dao"

func NewPropExecutor(key string, value interface{}) proposal.Executor {
	callback := func() error {
		if std.PrevRealm().PkgPath() != daoPkgPath {
			panic("should be executed from govdao")
		}
		std.SetParam(key, value)
		return nil
	}
	return proposal.NewExecutor(callback)
}
