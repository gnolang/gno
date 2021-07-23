package keys

import (
	"fmt"
	"path/filepath"
)

const defaultKeyDBName = "keys.db"

// NewKeyBaseFromDir initializes a keybase at a particular dir.
func NewKeyBaseFromDir(rootDir string) (Keybase, error) {
	return NewLazyDBKeybase(defaultKeyDBName, filepath.Join(rootDir, "keys")), nil
}

// NewInMemoryKeyBase returns a storage-less keybase.
func NewInMemoryKeyBase() Keybase { return NewInMemory() }

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
