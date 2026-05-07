package params

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// GasCostParamPerByte is the gas cost per byte of parameter value data.
// Go-side overhead is <1 ns/byte; store write costs provide primary DoS protection.
const GasCostParamPerByte int64 = 1

func X_setSysParamString(m *gno.Machine, module, submodule, name, val string) {
	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(overflow.Mulp(int64(len(val)), GasCostParamPerByte), "setSysParamString")
	}
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
	if m.GasMeter != nil {
		m.GasMeter.ConsumeGas(overflow.Mulp(int64(len(val)), GasCostParamPerByte), "setSysParamBytes")
	}
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetBytes(pk, val)
}

func X_setSysParamStrings(m *gno.Machine, module, submodule, name string, val []string) {
	if m.GasMeter != nil {
		total := int64(0)
		for _, s := range val {
			total += int64(len(s))
		}
		m.GasMeter.ConsumeGas(overflow.Mulp(total, GasCostParamPerByte), "setSysParamStrings")
	}
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.SetStrings(pk, val)
}

func X_updateSysParamStrings(m *gno.Machine, module, submodule, name string, val []string, add bool) {
	if m.GasMeter != nil {
		total := int64(0)
		for _, s := range val {
			total += int64(len(s))
		}
		m.GasMeter.ConsumeGas(overflow.Mulp(total, GasCostParamPerByte), "updateSysParamStrings")
	}
	assertSysParamsRealm(m)
	pk := prmkey(module, submodule, name)
	execctx.GetContext(m).Params.UpdateStrings(pk, val, add)
}

// G1: typed read helpers. Same realm gate as writes — only
// gno.land/r/sys/params can call these. The (value, found) shape
// lets callers distinguish "key never written" from "key written as
// zero/empty value." `found` comes directly from the keeper's
// existence check, so set-to-zero and unset are always reliably
// distinguishable for every type.

func X_getSysParamString(m *gno.Machine, module, submodule, name string) (string, bool) {
	assertSysParamsRealm(m)
	var out string
	ok := execctx.GetContext(m).Params.GetString(prmkey(module, submodule, name), &out)
	return out, ok
}

func X_getSysParamBool(m *gno.Machine, module, submodule, name string) (bool, bool) {
	assertSysParamsRealm(m)
	var out bool
	ok := execctx.GetContext(m).Params.GetBool(prmkey(module, submodule, name), &out)
	return out, ok
}

func X_getSysParamInt64(m *gno.Machine, module, submodule, name string) (int64, bool) {
	assertSysParamsRealm(m)
	var out int64
	ok := execctx.GetContext(m).Params.GetInt64(prmkey(module, submodule, name), &out)
	return out, ok
}

func X_getSysParamUint64(m *gno.Machine, module, submodule, name string) (uint64, bool) {
	assertSysParamsRealm(m)
	var out uint64
	ok := execctx.GetContext(m).Params.GetUint64(prmkey(module, submodule, name), &out)
	return out, ok
}

func X_getSysParamBytes(m *gno.Machine, module, submodule, name string) ([]byte, bool) {
	assertSysParamsRealm(m)
	var out []byte
	ok := execctx.GetContext(m).Params.GetBytes(prmkey(module, submodule, name), &out)
	return out, ok
}

func X_getSysParamStrings(m *gno.Machine, module, submodule, name string) ([]string, bool) {
	assertSysParamsRealm(m)
	var out []string
	ok := execctx.GetContext(m).Params.GetStrings(prmkey(module, submodule, name), &out)
	return out, ok
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
