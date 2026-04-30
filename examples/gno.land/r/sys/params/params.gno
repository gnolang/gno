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
//	pr := params.NewSysParamStringPropExecutor("bank", "p", "restricted_denoms")
package params

import (
	"chain"
	prms "sys/params"

	"gno.land/r/gov/dao"
)

// this is only used for emitting events.
func syskey(module, submodule, name string) string {
	return module + ":" + submodule + ":" + name
}

// assertNotValsetKey rejects governance proposals that target the
// node:valset:* key family. Those keys are reserved for the realm-side
// gate in r/sys/params/valset.gno (SetValsetProposal) which checks
// the immediate caller is gno.land/r/sys/validators/v3. Without this
// guard, a generic NewSysParam*PropRequest("node","valset",...) would
// let any GovDAO supermajority bypass the v3 authorization and write
// validator-set state directly.
func assertNotValsetKey(module, submodule string) {
	if module == "node" && submodule == "valset" {
		panic("node:valset:* is reserved for r/sys/validators/v3; use it instead of the generic factory")
	}
}

func NewSysParamStringPropRequest(module, submodule, name, value string) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamString(module, submodule, name, value) },
		"",
	)
}

func NewSysParamInt64PropRequest(module, submodule, name string, value int64) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamInt64(module, submodule, name, value) },
		"",
	)
}

func NewSysParamUint64PropRequest(module, submodule, name string, value uint64) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamUint64(module, submodule, name, value) },
		"",
	)
}

func NewSysParamBoolPropRequest(module, submodule, name string, value bool) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamBool(module, submodule, name, value) },
		"",
	)
}

func NewSysParamBytesPropRequest(module, submodule, name string, value []byte) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamBytes(module, submodule, name, value) },
		"",
	)
}

func NewSysParamStringsPropRequest(module, submodule, name string, value []string) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamStrings(module, submodule, name, value) },
		"",
	)
}

func NewSysParamStringsPropRequestWithTitle(module, submodule, name, title string, value []string) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.SetSysParamStrings(module, submodule, name, value) },
		title,
	)
}
func NewSysParamStringsPropRequestAddWithTitle(module, submodule, name, title string, value []string) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.UpdateSysParamStrings(module, submodule, name, value, true) },
		title,
	)
}
func NewSysParamStringsPropRequestRemoveWithTitle(module, submodule, name, title string, value []string) dao.ProposalRequest {
	assertNotValsetKey(module, submodule)
	return newPropRequest(
		syskey(module, submodule, name),
		func() { prms.UpdateSysParamStrings(module, submodule, name, value, false) },
		title,
	)
}
func newPropRequest(key string, fn func(), title string) dao.ProposalRequest {
	callback := func(cur realm) error {
		fn()
		chain.Emit("set", "key", key) // TODO document, make const, make consistent. 'k'??
		return nil
	}

	if title == "" {
		title = "Set new sys/params key"
	}

	e := dao.NewSimpleExecutor(callback, "")

	return dao.NewProposalRequest(title, "This proposal wants to add a new key to sys/params: "+key, e)
}
