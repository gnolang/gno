// Package params provides functions for creating parameter proposal requests that
// interface with the Params Keeper.
//
// This package enables setting various parameter types (such as strings,
// integers, booleans, and byte slices) through the GovDAO proposal mechanism.
// Each function returns an ProposalRequest that, when called, sets the specified
// parameter in the Params Keeper.
//
// The proposal requests are designed to be used within governance proposals to modify
// parameters dynamically. The integration with the GovDAO allows for parameter
// changes to be proposed and executed in a controlled manner, ensuring that
// modifications are subject to governance processes.
//
// Example usage:
//
//	pr := params.NewStringPropRequest("exampleKey", "exampleValue")
//	// This proposal request can be used in a governance proposal to create a proposal that will set a parameter if approved.
package params

import (
	"std"

	"gno.land/r/gov/dao/proxy"
)

func NewStringPropRequest(key string, value string) proxy.ProposalRequest {
	return newPropRequest(key, func() { std.SetParamString(key, value) })
}

func NewInt64PropRequest(key string, value int64) proxy.ProposalRequest {
	return newPropRequest(key, func() { std.SetParamInt64(key, value) })
}

func NewUint64PropRequest(key string, value uint64) proxy.ProposalRequest {
	return newPropRequest(key, func() { std.SetParamUint64(key, value) })
}

func NewBoolPropRequest(key string, value bool) proxy.ProposalRequest {
	return newPropRequest(key, func() { std.SetParamBool(key, value) })
}

func NewBytesPropRequest(key string, value []byte) proxy.ProposalRequest {
	return newPropRequest(key, func() { std.SetParamBytes(key, value) })
}

func newPropRequest(key string, fn func()) proxy.ProposalRequest {
	callback := func(proxy.Metadata) error {
		fn()
		std.Emit("set", "k", key)
		return nil
	}
	
	return proxy.ProposalRequest{
		Title:       "Set new sys/params key",
		Description: "This proposal wants to add a new key to sys/params: " + key,
		Callback:    callback,
	}
}
