package validators

import (
	"gno.land/p/sys/validators"
)

// GetChanges returns the validator changes stored on the realm, since the given block number.
// This function is unexported and intended to be called by gno.land through the GnoSDK
func GetChanges(from int64) []validators.Validator {
	// Construct the changes
	valsetChanges := make([]validators.Validator, 0)

	for _, change := range changes {
		if change.blockNum < from {
			continue
		}

		valsetChanges = append(valsetChanges, change.validator)
	}

	return valsetChanges
}
