package params

// SetParamXXX(k, v) are for setting arbitrary realm-local parameters that can be called from any realm.
// It may or may not affect system behavior.
func SetString(key string, val string)
func SetBool(key string, val bool)
func SetInt64(key string, val int64)
func SetUint64(key string, val uint64)
func SetBytes(key string, val []byte)
func SetStrings(key string, val []string)
