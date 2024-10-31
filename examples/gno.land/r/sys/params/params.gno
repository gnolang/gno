package params

import (
	"std"

	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
)

func NewStringPropExecutor(key string, value string) dao.Executor {
	return newPropExecutor(key, func() { std.SetParamString(key, value) })
}

func NewInt64PropExecutor(key string, value int64) dao.Executor {
	return newPropExecutor(key, func() { std.SetParamInt64(key, value) })
}

func NewUint64PropExecutor(key string, value uint64) dao.Executor {
	return newPropExecutor(key, func() { std.SetParamUint64(key, value) })
}

func NewBoolPropExecutor(key string, value bool) dao.Executor {
	return newPropExecutor(key, func() { std.SetParamBool(key, value) })
}

func NewBytesPropExecutor(key string, value []byte) dao.Executor {
	return newPropExecutor(key, func() { std.SetParamBytes(key, value) })
}

func newPropExecutor(key string, fn func()) dao.Executor {
	callback := func() error {
		fn()
		std.Emit("set", "k", key)
		return nil
	}
	return bridge.GovDAO().NewGovDAOExecutor(callback)
}
