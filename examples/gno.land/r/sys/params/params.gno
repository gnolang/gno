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

// assertNotValsetKey rejects generic proposals on node:valset:*, which only
// SetValsetProposal in valset.gno may write, for r/sys/validators/v0.
func assertNotValsetKey(module, submodule string) {
	if module == "node" && submodule == "valset" {
		panic("node:valset:* is reserved for r/sys/validators/v0; use it instead of the generic factory")
	}
}

// assertNotRunSubmittersKey rejects generic proposals on vm:p:run_submitters,
// which only ProposeSetRunSubmitters may set, with its own checks.
func assertNotRunSubmittersKey(module, submodule, name string) {
	if module == "vm" && submodule == "p" && name == runSubmittersKey {
		panic(runSubmittersKey + " is reserved for ProposeSetRunSubmitters in this realm; " +
			"use it instead of the generic factory")
	}
}

func NewSysParamStringPropRequest(cur realm, module, submodule, name, value string) dao.ProposalRequest {
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamString(module, submodule, name, value) },
		"",
	)
}

func NewSysParamInt64PropRequest(cur realm, module, submodule, name string, value int64) dao.ProposalRequest {
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamInt64(module, submodule, name, value) },
		"",
	)
}

func NewSysParamUint64PropRequest(cur realm, module, submodule, name string, value uint64) dao.ProposalRequest {
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamUint64(module, submodule, name, value) },
		"",
	)
}

func NewSysParamBoolPropRequest(cur realm, module, submodule, name string, value bool) dao.ProposalRequest {
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamBool(module, submodule, name, value) },
		"",
	)
}

func NewSysParamBytesPropRequest(cur realm, module, submodule, name string, value []byte) dao.ProposalRequest {
	value = copyBytes(value)
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamBytes(module, submodule, name, value) },
		"",
	)
}

func NewSysParamStringsPropRequest(cur realm, module, submodule, name string, value []string) dao.ProposalRequest {
	return NewSysParamStringsPropRequestWithTitle(cur, module, submodule, name, "", value)
}

func NewSysParamStringsPropRequestWithTitle(cur realm, module, submodule, name, title string, value []string) dao.ProposalRequest {
	value = copyStrings(value)
	return newPropRequest(cur, module, submodule, name,
		func() { prms.SetSysParamStrings(module, submodule, name, value) },
		title,
	)
}
func NewSysParamStringsPropRequestAddWithTitle(cur realm, module, submodule, name, title string, value []string) dao.ProposalRequest {
	return newStringsUpdatePropRequest(cur, module, submodule, name, title, value, true)
}
func NewSysParamStringsPropRequestRemoveWithTitle(cur realm, module, submodule, name, title string, value []string) dao.ProposalRequest {
	return newStringsUpdatePropRequest(cur, module, submodule, name, title, value, false)
}
func newStringsUpdatePropRequest(cur realm, module, submodule, name, title string, value []string, add bool) dao.ProposalRequest {
	value = copyStrings(value)
	return newPropRequest(cur, module, submodule, name,
		func() { prms.UpdateSysParamStrings(module, submodule, name, value, add) },
		title,
	)
}
func newPropRequest(cur realm, module, submodule, name string, fn func(), title string) dao.ProposalRequest {
	// Every generic factory goes through here, so none can skip these checks.
	assertNotValsetKey(module, submodule)
	assertNotRunSubmittersKey(module, submodule, name)

	key := syskey(module, submodule, name)
	callback := func(cur realm) error {
		fn()
		chain.Emit("set", "key", key) // TODO document, make const, make consistent. 'k'??
		return nil
	}

	if title == "" {
		title = "Set new sys/params key"
	}

	e := dao.NewSimpleExecutor(0, cur, callback, "")

	return dao.NewProposalRequest(title, "This proposal wants to add a new key to sys/params: "+key, e)
}

// copyBytes returns src in a fresh array this realm owns.
func copyBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	cp := make([]byte, len(src))
	copy(cp, src)
	return cp
}

// copyStrings returns src in a fresh array this realm owns.
func copyStrings(src []string) []string {
	if src == nil {
		return nil
	}
	cp := make([]string, len(src))
	copy(cp, src)
	return cp
}
