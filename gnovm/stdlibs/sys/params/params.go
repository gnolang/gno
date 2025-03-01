package params

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/std"
)

func X_setSysParamString(m *gno.Machine, module, realm, key, val string) {
	pk := paramskey(module, realm, key)
	std.GetContext(m).Params.SetString(pk, val)
}

func X_setSysParamBool(m *gno.Machine, module, realm, key string, val bool) {
	pk := paramskey(module, realm, key)
	std.GetContext(m).Params.SetBool(pk, val)
}

func X_setSysParamInt64(m *gno.Machine, module, realm, key string, val int64) {
	pk := paramskey(module, realm, key)
	std.GetContext(m).Params.SetInt64(pk, val)
}

func X_setSysParamUint64(m *gno.Machine, module, realm, key string, val uint64) {
	pk := paramskey(module, realm, key)
	std.GetContext(m).Params.SetUint64(pk, val)
}

func X_setSysParamBytes(m *gno.Machine, module, realm, key string, val []byte) {
	pk := paramskey(module, realm, key)
	std.GetContext(m).Params.SetBytes(pk, val)
}

func paramskey(module, realm, key string) string {
	if realm != "" {
		return module + ":" + realm + ":" + key
	}
	return module + ":" + key
}
