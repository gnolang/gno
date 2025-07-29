package haystack

import (
	phaystack "gno.land/p/n2p5/haystack"
)

// ResetStorage resets the storage for testing
func ResetStorage() {
	storage = phaystack.New()
}
