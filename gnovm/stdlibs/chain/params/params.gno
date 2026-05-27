// Package params provides a set of functions for setting arbitrary realm-local
// parameters that can be called from any realm.
// It may or may not affect system behavior.
package params

func SetString(key string, val string)
func SetBool(key string, val bool)
func SetInt64(key string, val int64)
func SetUint64(key string, val uint64)
func SetBytes(key string, val []byte)
func SetStrings(key string, val []string)
func UpdateParamStrings(key string, val []string, add bool)

// SetBytesKey stores val under a raw binary key. Unlike SetBytes, the key may
// contain any bytes, including ':'.
func SetBytesKey(key []byte, val []byte)

// GetBytesKey reads back a value stored with SetBytesKey. It returns
// (value, true) when the realm-local key is set, and (nil, false) otherwise.
func GetBytesKey(key []byte) ([]byte, bool)
