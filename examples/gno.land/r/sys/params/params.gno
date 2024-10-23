package params

import (
	"std"

	"gno.land/p/gov/proposal"
)

func NewPropExecutor(key string, value interface{}) proposal.Executor {
	callback := func() error {
		assertGovDAOCaller()
		std.SetParam(key, value)
		return nil
	}

	return proposal.NewExecutor(callback)
}

const daoPkgPath = "gno.land/r/gov/dao"

func assertGovDAOCaller() {
	if std.PrevRealm().PkgPath() != daoPkgPath {
		panic(errNotGovDAO)
	}
}
