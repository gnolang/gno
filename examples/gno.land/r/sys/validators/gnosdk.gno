package validators

import "gno.land/p/sys/validators"

// getChanges returns the validator changes stored on the realm.
// This function is unexported and intended to be called by gno.land through the GnoSDK
func getChanges() []validators.Validator {
	// Construct the changes
	valsetChanges := make([]validators.Validator, len(changes))
	copy(valsetChanges, changes)

	// Reset the changes set
	changes = changes[:0]

	return valsetChanges
}
