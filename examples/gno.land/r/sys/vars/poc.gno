package vars

import (
	"std"

	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

var errNotGovDAO = "caller not govdao executor"

type KV struct {
	Key   string
	Value interface{}
}

// NewVarPropExecutor creates a new proposal executor for
// changing r/sys/vars values through governance proposals
func NewVarPropExecutor(changesFn func() []KV) dao.Executor {
	cb := func() error {
		assertGovDAOCaller()

		for _, kv := range changesFn() {
			if kv.Value == nil {
				// Removal request
				vars.Remove(kv.Key)

				continue
			}

			// Add / update request
			vars.Set(kv.Key, kv.Value)
		}

		return nil
	}

	return executor.NewCallbackExecutor(cb)
}

// assertGovDAOCaller verifies the caller is the GovDAO executor
func assertGovDAOCaller() {
	govdaoPath := GetStringValue(GovdaoRealm)

	if std.PrevRealm().PkgPath() != govdaoPath {
		panic(errNotGovDAO)
	}
}
