package std

func setParamString(key string, val string)
func setParamBool(key string, val bool)
func setParamInt64(key string, val int64)
func setParamUint64(key string, val uint64)
func setParamBytes(key string, val []byte)

var (
	SetParamString = setParamString
	SetParamBool   = setParamBool
	SetParamInt64  = setParamInt64
	SetParamUint64 = setParamUint64
	SetParamBytes  = setParamBytes
)
