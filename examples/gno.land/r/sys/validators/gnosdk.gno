package validators

import "gno.land/p/sys/validators"

// getChanges returns the validator changes stored on the realm.
// This function is unexported and intended to be called by gno.land through the GnoSDK
func getChanges() []validators.Validator {
	// Construct the changes
	changes := make([]validators.Validator, len(v.changes))
	copy(changes, v.changes)

	// Reset the changes set
	v.changes = v.changes[:0]

	return changes
}
