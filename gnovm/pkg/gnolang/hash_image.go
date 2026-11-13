package gnolang

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
)

type ValueHash struct {
	Hashlet
}

func (vh ValueHash) MarshalAmino() (string, error) {
	return hex.EncodeToString(vh.Hashlet[:]), nil
}

func (vh *ValueHash) UnmarshalAmino(h string) error {
	_, err := hex.Decode(vh.Hashlet[:], []byte(h))
	return err
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

// IsZero reports whether h is all zeros. Hand-unrolled into word
// loads: the generic comparison against a zero [20]byte lowers to a
// runtime.memequal call, and IsZero sits on per-write VM hot paths
// (ObjectID.IsZero in the readonly/cross-realm guards).
func (h Hashlet) IsZero() bool {
	return binary.NativeEndian.Uint64(h[0:8])|
		binary.NativeEndian.Uint64(h[8:16])|
		uint64(binary.NativeEndian.Uint32(h[16:20])) == 0
}

func HashBytes(bz []byte) (res Hashlet) {
	hash := sha256.Sum256(bz)
	copy(res[:], hash[:HashSize])
	return
}
