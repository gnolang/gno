package vars

import (
	"errors"
	"strings"

	"gno.land/p/demo/avl"
)

// maxRequestKeys is the maximum number of keys
// that can be returned on request
const maxRequestKeys = 50

// Prefixes

// GovernancePrefix is used for goveranance-related vars
const GovernancePrefix = "gov-"

// Initial keys

const GovdaoRealm = GovernancePrefix + "govdao-realm"

var (
	ErrMissingValue     = errors.New("missing value")
	ErrInvalidValueType = errors.New("invalid value type")
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
			GovdaoRealm,
			"gno.land/r/gov/dao",
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

// GetKeysWithPrefix returns the paginated key values with the specified prefix
func GetKeysWithPrefix(
	prefix string,
	offset,
	count uint64,
) []string {
	// Calculate the left and right bounds
	if count < 1 || offset >= uint64(vars.Size()) {
		return []string{}
	}

	// Limit the maximum number of returned keys
	if count > maxRequestKeys {
		count = maxRequestKeys
	}

	var (
		keys       = make([]string, 0)
		currOffset = uint64(0)
		currCount  = uint64(0)
	)

	vars.Iterate("", "", func(key string, _ interface{}) bool {
		// Skip over specific elements
		if currOffset < offset {
			currOffset++

			return false
		}

		// Check if the key matches
		if !strings.HasPrefix(key, prefix) {
			return false
		}

		// Save the key
		keys = append(keys, key)

		// Check if the count elements
		// have been gathered
		currCount++
		return currCount == count
	})

	return keys
}
