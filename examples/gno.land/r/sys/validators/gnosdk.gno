package validators

import (
	"gno.land/p/sys/validators"
)

// GetChanges returns the validator changes stored on the realm, since the given block number.
// This function is intended to be called by gno.land through the GnoSDK
func GetChanges(from int64) []validators.Validator {
	valsetChanges := make([]validators.Validator, 0)

	// Gather the changes from the specified block
	changes.Iterate(getBlockID(from), "", func(_ string, value interface{}) bool {
		chs := value.([]change)

		if len(chs) > 0 && chs[0].blockNum < from {
			// Each change entry in the set is tied to the same block number,
			// so there is no need to parse other changes
			// if the block number is invalid (< from)
			return false
		}

		for _, ch := range chs {
			valsetChanges = append(valsetChanges, ch.validator)
		}

		return false
	})

	return valsetChanges
}
