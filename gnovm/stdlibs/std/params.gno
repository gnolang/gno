package std

// These are native bindings to the banker's functions.
func setConfigString(key string, val string)
func setConfigBool(key string, val bool)
func setConfigInt64(key string, val int64)
func setConfigUint64(key string, val uint64)

// XXX: add doc
func SetConfig(key string, val interface{}) {
	switch v := val.(type) {
	case string:
		setConfigString(key, v)
	case bool:
		setConfigBool(key, v)
	case int64:
		setConfigInt64(key, v)
	case uint64:
		setConfigUint64(key, v)
	default:
		panic("unsupported type")
	}
}

// XXX: add doc
//func GetConfig(key string) string {
//	return getConfig(key)
//}
