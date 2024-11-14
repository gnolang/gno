// Package params provides functions for creating parameter executors that
// interface with the Params Keeper.
//
// This package enables setting various parameter types (such as strings,
// integers, booleans, and byte slices) through the GovDAO proposal mechanism.
// Each function returns an executor that, when called, sets the specified
// parameter in the Params Keeper.
//
// The executors are designed to be used within governance proposals to modify
// parameters dynamically. The integration with the GovDAO allows for parameter
// changes to be proposed and executed in a controlled manner, ensuring that
// modifications are subject to governance processes.
//
// Example usage:
//
//	executor := params.NewStringPropExecutor("exampleKey", "exampleValue")
//	// This executor can be used in a governance proposal to set the parameter.
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
