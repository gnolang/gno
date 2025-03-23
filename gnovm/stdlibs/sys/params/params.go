package params

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_setSysParamString(m *gno.Machine, module, submodule, name, val string) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetString(pk, val)
}

func X_setSysParamBool(m *gno.Machine, module, submodule, name string, val bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetBool(pk, val)
}

func X_setSysParamInt64(m *gno.Machine, module, submodule, name string, val int64) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetInt64(pk, val)
}

func X_setSysParamUint64(m *gno.Machine, module, submodule, name string, val uint64) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetUint64(pk, val)
}

func X_setSysParamBytes(m *gno.Machine, module, submodule, name string, val []byte) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetBytes(pk, val)
}

func X_setSysParamStrings(m *gno.Machine, module, submodule, name string, val []string) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	std.GetContext(m).Params.SetStrings(pk, val)
}

// @moul, just a note on this, because I think it's important:
// It seems overly restrictive to have the sys/params realm be the only
// entry point for sys params changes, given its limited API. It's not easy
// to "bundle" logic with params changes, since if we centralize it to sys/params
// as it is now, we enforce a single request-single param change policy, with no arbitrary
// logic to go with the change. TL;DR we can't bundle additional logic in the GovDAO callback with
// a sys param change
func assertSysParamsRealm(m *gno.Machine) {
	// XXX improve
	if len(m.Frames) < 2 {
		panic("should not happen")
	}
	if m.Frames[len(m.Frames)-1].LastPackage.PkgPath != "sys/params" {
		panic("should not happen")
	}
	if m.Frames[len(m.Frames)-2].LastPackage.PkgPath != "gno.land/r/sys/params" {
		// XXX this should not happen after import rule.
		panic(`"sys/params" can only be used from "gno.land/r/sys/params"`)
	}
}

func prmkey(module, submodule, name string) string {
	// XXX consolidate validation
	if strings.Contains(name, ":") {
		panic("invalid param name: " + name)
	}
	if submodule == "" {
		panic("submodule cannot be empty")
	}
	return module + ":" + submodule + ":" + name
}
