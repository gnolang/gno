package params

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
)

func X_setSysParamString(m *gno.Machine, module, submodule, name, val string) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetString(pk, val)
}

func X_setSysParamBool(m *gno.Machine, module, submodule, name string, val bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetBool(pk, val)
}

func X_setSysParamInt64(m *gno.Machine, module, submodule, name string, val int64) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetInt64(pk, val)
}

func X_setSysParamUint64(m *gno.Machine, module, submodule, name string, val uint64) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetUint64(pk, val)
}

func X_setSysParamBytes(m *gno.Machine, module, submodule, name string, val []byte) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetBytes(pk, val)
}

func X_setSysParamStrings(m *gno.Machine, module, submodule, name string, val []string) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetStrings(pk, val)
}

func X_updateSysParamStrings(m *gno.Machine, module, submodule, name string, val []string, add bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.UpdateStrings(pk, val, add)
}

// G1: typed read helpers. Same realm gate as writes — only
// gno.land/r/sys/params can call these. The (value, found) shape lets
// callers like v3.init() distinguish "key never written" from
// "key written as zero/empty value".

func X_getSysParamString(m *gno.Machine, module, submodule, name string) (string, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	var out string
	execctx.GetContext(m).Params.GetString(pk, &out)
	return out, out != ""
}

func X_getSysParamBool(m *gno.Machine, module, submodule, name string) (bool, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	// Bool's zero value is false; we cannot distinguish "unset" from
	// "set to false" via the keeper's getIfExists. Callers that need
	// this distinction should use a string/bytes key instead. For
	// API symmetry, return (val, true) — i.e., always-found.
	var out bool
	execctx.GetContext(m).Params.GetBool(pk, &out)
	return out, true
}

func X_getSysParamInt64(m *gno.Machine, module, submodule, name string) (int64, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	var out int64
	execctx.GetContext(m).Params.GetInt64(pk, &out)
	return out, true // see GetBool comment re: zero-vs-unset.
}

func X_getSysParamUint64(m *gno.Machine, module, submodule, name string) (uint64, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	var out uint64
	execctx.GetContext(m).Params.GetUint64(pk, &out)
	return out, true
}

func X_getSysParamBytes(m *gno.Machine, module, submodule, name string) ([]byte, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	var out []byte
	execctx.GetContext(m).Params.GetBytes(pk, &out)
	return out, out != nil
}

func X_getSysParamStrings(m *gno.Machine, module, submodule, name string) ([]string, bool) {
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	var out []string
	execctx.GetContext(m).Params.GetStrings(pk, &out)
	return out, out != nil
}

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
