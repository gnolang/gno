package validators

import (
	"sort"

	"gno.land/p/sys/validators"
)

// GetChanges returns the validator changes stored on the realm, since the given block number.
// This function is unexported and intended to be called by gno.land through the GnoSDK
func GetChanges(from int64) []validators.Validator {
	// Perform binary search to find the starting index
	idx := sort.Search(len(changes), func(i int) bool {
		return changes[i].blockNum >= from
	})

	// Construct the changes from the found index
	valsetChanges := make([]validators.Validator, 0)
	for i := idx; i < len(changes); i++ {
		valsetChanges = append(valsetChanges, changes[i].validator)
	}

	return valsetChanges
}
