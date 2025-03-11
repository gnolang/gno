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
//	// This executor can be used in a governance proposal to set the parameter.
//	executor := params.NewStringsPropExecutor("bank", "p", "restricted_denoms")
package params

import (
	"std"
	prms "sys/params"

	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
)

// this is only used for emitting events.
func syskey(module, submodule, name string) string {
	return module + ":" + submodule + ":" + name
}

func NewSysParamStringPropExecutor(module, submodule, name string, value string) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamString(module, submodule, name, value) },
	)
}

func NewSysParamInt64PropExecutor(module, submodule, name string, value int64) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamInt64(module, submodule, name, value) },
	)
}

func NewSysParamUint64PropExecutor(module, submodule, name string, value uint64) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamUint64(module, submodule, name, value) },
	)
}

func NewSysParamBoolPropExecutor(module, submodule, name string, value bool) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamBool(module, submodule, name, value) },
	)
}

func NewSysParamBytesPropExecutor(module, submodule, name string, value []byte) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamBytes(module, submodule, name, value) },
	)
}

func NewSysParamStringsPropExecutor(module, submodule, name string, value []string) dao.Executor {
	return newPropExecutor(
		syskey(module, submodule, name),
		func() { prms.SetSysParamStrings(module, submodule, name, value) },
	)
}

func newPropExecutor(key string, fn func()) dao.Executor {
	callback := func() error {
		fn()
		std.Emit("set", "key", key) // TODO document, make const, make consistent. 'k'??
		return nil
	}
	return bridge.GovDAO().NewGovDAOExecutor(callback)
}
func propose(exec dao.Executor, title, desc string) uint64 {
	// The executor's callback function is executed only after the proposal
	// has been voted on and approved by the GovDAO.
	prop := dao.ProposalRequest{
		Title:       title,
		Description: desc,
		Executor:    exec,
	}
	return bridge.GovDAO().Propose(prop)
}
