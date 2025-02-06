package local

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// LocalSigner implements types.Signer using a FileKey persisted to disk.
type LocalSigner struct {
	key *FileKey
}

// LocalSigner type implements types.Signer.
var _ types.Signer = (*LocalSigner)(nil)

// PubKey implements types.Signer.
func (fs *LocalSigner) PubKey() (crypto.PubKey, error) {
	return fs.key.PubKey, nil
}

// Sign implements types.Signer.
func (fs *LocalSigner) Sign(signBytes []byte) ([]byte, error) {
	return fs.key.PrivKey.Sign(signBytes)
}

// LocalSigner type implements fmt.Stringer.
var _ fmt.Stringer = (*LocalSigner)(nil)

// String implements fmt.Stringer.
func (fs *LocalSigner) String() string {
	return fmt.Sprintf("LocalSigner(%v)", fs.key.Address)
}

// NewLocalSigner returns a new LocalSigner instance using a file key from the given
// file path. If the file does not exist, a new random FileKey is generated and
// persisted to disk.
func NewLocalSigner(filePath string) (*LocalSigner, error) {
	// Load existing file key or generate a new random one.
	key, err := NewFileKey(filePath)
	if err != nil {
		return nil, err
	}

	return &LocalSigner{key}, nil
}
