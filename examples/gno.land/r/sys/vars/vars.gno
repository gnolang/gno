package vars

import (
	"errors"

	"gno.land/p/demo/avl"
)

// maxRequestKeys is the maximum number of keys
// that can be returned on request
const maxRequestKeys = 50

// Prefixes

// SystemPrefix is used for system-related vars
const SystemPrefix = "sys-"

// Initial keys

const (
	ValidatorsRealm     = SystemPrefix + "validators-realm"
	ValidatorsChangesFn = SystemPrefix + "validators-changes-fn"
)

var (
	ErrMissingValue     = errors.New("missing value")
	ErrInvalidValueType = errors.New("invalid value type")
	ErrTooManyValues    = errors.New("too many values requested")
)

// vars is a simple KV store
var vars *avl.Tree

func init() {
	// Initial KV set
	initialSet := []struct {
		key   string
		value string
	}{
		{
			ValidatorsRealm,
			"gno.land/r/sys/validators/v2",
		},
		{
			ValidatorsChangesFn,
			"GetChanges",
		},
	}

	// Save the initial KV set
	vars = avl.NewTree()

	for _, v := range initialSet {
		vars.Set(v.key, v.value)
	}
}

// GetValue fetches the given value from the store, if any
func GetValue(key string) interface{} {
	val, exists := vars.Get(key)
	if !exists {
		panic(ErrMissingValue)
	}

	return val
}

// GetStringValue returns the string value associated
// with the given key, if any
func GetStringValue(key string) string {
	// Get the value
	val := GetValue(key)

	// Make sure it's actually a string
	valStr, ok := val.(string)
	if !ok {
		panic(ErrInvalidValueType)
	}

	return valStr
}

// GetStringValues fetches a list of values
func GetStringValues(keys ...string) []string {
	if len(keys) > maxRequestKeys {
		panic(ErrTooManyValues)
	}

	vals := make([]string, 0, len(keys))
	for _, key := range keys {
		vals = append(vals, GetStringValue(key))
	}

	return vals
}

// GetKeys returns the paginated key values
func GetKeys(offset, count uint64) []string {
	// Calculate the left and right bounds
	if count < 1 || offset >= uint64(vars.Size()) {
		return []string{}
	}

	// Limit the maximum number of returned keys
	if count > maxRequestKeys {
		count = maxRequestKeys
	}

	// Gather the members
	keys := make([]string, 0)
	vars.IterateByOffset(
		int(offset),
		int(count),
		func(key string, _ interface{}) bool {
			// Save the key
			keys = append(keys, key)

			return false
		},
	)

	return keys
}
