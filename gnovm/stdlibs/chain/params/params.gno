// Package params provides a set of functions for setting and reading arbitrary
// realm-local parameters that can be called from any realm.
// It may or may not affect system behavior.
package params

func SetString(key string, val string)
func SetBool(key string, val bool)
func SetInt64(key string, val int64)
func SetUint64(key string, val uint64)
func SetBytes(key string, val []byte)
func SetStrings(key string, val []string)
func UpdateParamStrings(key string, val []string, add bool)

// GetXxx reads back a value previously stored under key with the matching
// SetXxx. It returns (value, true) when the realm-local key is set, and
// (zero, false) otherwise.
func GetString(key string) (string, bool)
func GetBool(key string) (bool, bool)
func GetInt64(key string) (int64, bool)
func GetUint64(key string) (uint64, bool)
func GetBytes(key string) ([]byte, bool)
func GetStrings(key string) ([]string, bool)
