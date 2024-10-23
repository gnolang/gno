package params

import (
	"std"

	"gno.land/p/gov/proposal"
)

const daoPkgPath = "gno.land/r/gov/dao"

func NewStringPropExecutor(key string, value string) proposal.Executor {
	return newPropExecutor(func() { std.SetParamString(key, value) })
}

func NewInt64PropExecutor(key string, value int64) proposal.Executor {
	return newPropExecutor(func() { std.SetParamInt64(key, value) })
}

func NewUint64PropExecutor(key string, value uint64) proposal.Executor {
	return newPropExecutor(func() { std.SetParamUint64(key, value) })
}

func NewBoolPropExecutor(key string, value bool) proposal.Executor {
	return newPropExecutor(func() { std.SetParamBool(key, value) })
}

func NewBytesPropExecutor(key string, value []byte) proposal.Executor {
	return newPropExecutor(func() { std.SetParamBytes(key, value) })
}

func newPropExecutor(fn func()) proposal.Executor {
	callback := func() error {
		if std.PrevRealm().PkgPath() != daoPkgPath {
			panic("should be executed from govdao")
		}
		fn()
		return nil
	}
	return proposal.NewExecutor(callback)
}
