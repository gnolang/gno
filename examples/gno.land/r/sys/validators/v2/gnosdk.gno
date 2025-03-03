package validators

import (
	"gno.land/p/sys/validators"
)

// GetChanges returns the validator changes stored on the realm, since the given block number.
// This function is intended to be called by gno.land through the GnoSDK
func GetChanges(from int64) []validators.Validator {
	valsetChanges := make([]validators.Validator, 0)

	// Gather the changes from the specified block
	changes.Iterate(getBlockID(from), "", func(_ string, value any) bool {
		chs := value.([]change)

		for _, ch := range chs {
			valsetChanges = append(valsetChanges, ch.validator)
		}

		return false
	})

	return valsetChanges
}
