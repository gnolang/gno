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
	prms "sys/params"

	"gno.land/p/demo/dao"
	"gno.land/r/gov/dao/bridge"
)

func syskey(module, realm, key string) string {
	if realm == "" {
		return module + ":" + key
	} else {
		return module + ":" + realm + ":" + key
	}
}

func NewSysParamStringPropExecutor(module, realm, key string, value string) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamString(module, realm, key, value) },
	)
}

func NewSysParamInt64PropExecutor(module, realm, key string, value int64) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamInt64(module, realm, key, value) },
	)
}

func NewSysParamUint64PropExecutor(module, realm, key string, value uint64) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamUint64(module, realm, key, value) },
	)
}

func NewSysParamBoolPropExecutor(module, realm, key string, value bool) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamBool(module, realm, key, value) },
	)
}

func NewSysParamBytesPropExecutor(module, realm, key string, value []byte) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamBytes(module, realm, key, value) },
	)
}

func NewSysParamStringsPropExecutor(module, realm, key string, value []string) dao.Executor {
	return newPropExecutor(
		syskey(module, realm, key),
		func() { prms.SetSysParamStrings(module, realm, key, value) },
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
