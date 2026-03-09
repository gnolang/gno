package params

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

//XXX: benchmark the real cost
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
