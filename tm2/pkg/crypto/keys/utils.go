package keys

import (
	"fmt"
	"path/filepath"
)

const (
	defaultKeyDBName = "keys"
	defaultKeyDBDir  = "data"
)

// NewKeyBaseFromDir initializes a keybase at a particular dir.
func NewKeyBaseFromDir(rootDir string) (Keybase, error) {
	return NewLazyDBKeybase(defaultKeyDBName, filepath.Join(rootDir, defaultKeyDBDir)), nil
}

func ValidateMultisigThreshold(k, nKeys int) error {
	if k <= 0 {
		return fmt.Errorf("threshold must be a positive integer")
	}
	if nKeys < k {
		return fmt.Errorf(
			"threshold k of n multisignature: %d < %d", nKeys, k)
	}
	return nil
}
