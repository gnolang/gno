package std

// These are native bindings to the banker's functions.
func setParamString(key string, val string)
func setParamBool(key string, val bool)
func setParamInt64(key string, val int64)
func setParamUint64(key string, val uint64)
func setParamBytes(key string, val []byte)

// XXX: add doc
func SetParam(key string, val interface{}) {
	switch v := val.(type) {
	case string:
		setParamString(key, v)
	case bool:
		setParamBool(key, v)
	case int64:
		setParamInt64(key, v)
	case uint64:
		setParamUint64(key, v)
	case []byte:
		setParamBytes(key, v)
	default:
		panic("unsupported type")
	}
}

//func GetConfig(key string) string {
//	return getConfig(key)
//}
