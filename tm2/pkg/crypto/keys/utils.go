package keys

import (
	"fmt"
	"path/filepath"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
)

const defaultKeyDBName = "keys"

// NewKeyBaseFromDir initializes a keybase at a particular dir.
func NewKeyBaseFromDir(rootDir string) (Keybase, error) {
	return NewLazyDBKeybase(defaultKeyDBName, filepath.Join(rootDir, config.DefaultDBDir)), nil
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
