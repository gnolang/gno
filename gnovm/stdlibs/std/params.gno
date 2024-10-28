package std

func setParamString(key string, val string)
func setParamBool(key string, val bool)
func setParamInt64(key string, val int64)
func setParamUint64(key string, val uint64)
func setParamBytes(key string, val []byte)

func SetParamString(key string, val string) { setParamString(key, val) }
func SetParamBool(key string, val bool)     { setParamBool(key, val) }
func SetParamInt64(key string, val int64)   { setParamInt64(key, val) }
func SetParamUint64(key string, val uint64) { setParamUint64(key, val) }
func SetParamBytes(key string, val []byte)  { setParamBytes(key, val) }
