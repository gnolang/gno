package params

// SetSysParam*(module, k, v) is used for directly setting params in
// ExecContext.Params, or SysParam.
// Access to "sys/params" must be restricted.

func SetSysParamString(module, realm, key string, val string) {
	setSysParamString(module, realm, key, val)
}

func SetSysParamBool(module, realm, key string, val bool) {
	setSysParamBool(module, realm, key, val)
}

func SetSysParamInt64(module, realm, key string, val int64) {
	setSysParamInt64(module, realm, key, val)
}

func SetSysParamUint64(module, realm, key string, val uint64) {
	setSysParamUint64(module, realm, key, val)
}

func SetSysParamBytes(module, realm, key string, val []byte) {
	setSysParamBytes(module, realm, key, val)
}

func SetSysParamStrings(module, realm, key string, val []string) {
	if val == nil {
		val = []string{}
	}
	setSysParamStrings(module, realm, key, val)
}

func setSysParamString(module, realm, key string, val string)
func setSysParamBool(module, realm, key string, val bool)
func setSysParamInt64(module, realm, key string, val int64)
func setSysParamUint64(module, realm, key string, val uint64)
func setSysParamBytes(module, realm, key string, val []byte)
func setSysParamStrings(module, realm, key string, val []string)
