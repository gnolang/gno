package params

// SetSysParam*(module, submodule, name, value) is used for directly setting params in
// ExecContext.Params, or SysParam.
// Access to "sys/params" must be restricted.

func SetSysParamString(module, submodule, name string, val string) {
	setSysParamString(module, submodule, name, val)
}

func SetSysParamBool(module, submodule, name string, val bool) {
	setSysParamBool(module, submodule, name, val)
}

func SetSysParamInt64(module, submodule, name string, val int64) {
	setSysParamInt64(module, submodule, name, val)
}

func SetSysParamUint64(module, submodule, name string, val uint64) {
	setSysParamUint64(module, submodule, name, val)
}

func SetSysParamBytes(module, submodule, name string, val []byte) {
	setSysParamBytes(module, submodule, name, val)
}

func SetSysParamStrings(module, submodule, name string, val []string) {
	if val == nil {
		val = []string{}
	}
	setSysParamStrings(module, submodule, name, val)
}

func UpdateSysParamStrings(module, submodule, name string, val []string, add bool) {
	if val == nil {
		val = []string{}
	}
	updateSysParamStrings(module, submodule, name, val, add)
}

func setSysParamString(module, submodule, name string, val string)
func setSysParamBool(module, submodule, name string, val bool)
func setSysParamInt64(module, submodule, name string, val int64)
func setSysParamUint64(module, submodule, name string, val uint64)
func setSysParamBytes(module, submodule, name string, val []byte)
func setSysParamStrings(module, submodule, name string, val []string)
func updateSysParamStrings(module, submodule, name string, val []string, add bool)
