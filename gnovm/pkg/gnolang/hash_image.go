package gnolang

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

type ValueHash struct {
	Hashlet
}

func (vh ValueHash) MarshalAmino() (string, error) {
	return hex.EncodeToString(vh.Hashlet[:]), nil
}

func (vh *ValueHash) UnmarshalAmino(h string) error {
	return vh.Hashlet.UnmarshalHex(h)
}

func (vh ValueHash) Copy() ValueHash {
	return ValueHash{vh.Hashlet.Copy()}
}

//----------------------------------------
// Hash*

const HashSize = 20

type Hashlet [HashSize]byte

func NewHashlet(bz []byte) Hashlet {
	res := Hashlet{}
	if len(bz) != HashSize {
		panic("invalid input size")
	}
	copy(res[:], bz)
	return res
}

func (h Hashlet) Copy() Hashlet {
	return h
}

func (h Hashlet) Bytes() []byte {
	return h[:]
}

// UnmarshalHex decodes s into h, refusing any s that is not exactly HashSize
// bytes of hex, since hex.Decode panics past the array and pads short input.
func (h *Hashlet) UnmarshalHex(s string) error {
	if len(s) != hex.EncodedLen(HashSize) {
		return errors.New("invalid hash length: expected %d hex characters, got %d", hex.EncodedLen(HashSize), len(s))
	}
	_, err := hex.Decode(h[:], []byte(s))
	return err
}

func (h Hashlet) IsZero() bool {
	return h == Hashlet{}
}

func HashBytes(bz []byte) (res Hashlet) {
	hash := sha256.Sum256(bz)
	copy(res[:], hash[:HashSize])
	return
}
