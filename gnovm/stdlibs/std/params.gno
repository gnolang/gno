package std

func setParamString(key string, val string)
func setParamBool(key string, val bool)
func setParamInt64(key string, val int64)
func setParamUint64(key string, val uint64)
func setParamBytes(key string, val []byte)
func setParamStrings(key string, val []string)

func getParamString(key string) (string, bool)
func getParamBool(key string) (bool, bool)
func getParamInt64(key string) (int64, bool)
func getParamUint64(key string) (uint64, bool)
func getParamBytes(key string) ([]byte, bool)
func getParamStrings(key string) ([]string, bool)

// SetParamXXX(k, v) are for setting arbitrary realm-local parameters that can be called from any realm.
// It may or may not affect system behavior.
func SetParamString(key string, val string)    { setParamString(key, val) }
func SetParamBool(key string, val bool)        { setParamBool(key, val) }
func SetParamInt64(key string, val int64)      { setParamInt64(key, val) }
func SetParamUint64(key string, val uint64)    { setParamUint64(key, val) }
func SetParamBytes(key string, val []byte)     { setParamBytes(key, val) }
func SetParamStrings(key string, val []string) { setParamStrings(key, val) }

// GetParamXXX(k) are for getting arbitrary realm-local parameters that can be called from any realm
func GetParamString(key string) (string, bool)    { return getParamString(key) }
func GetParamBool(key string) (bool, bool)        { return getParamBool(key) }
func GetParamInt64(key string) (int64, bool)      { return getParamInt64(key) }
func GetParamUint64(key string) (uint64, bool)    { return getParamUint64(key) }
func GetParamBytes(key string) ([]byte, bool)     { return getParamBytes(key) }
func GetParamStrings(key string) ([]string, bool) { return getParamStrings(key) }
