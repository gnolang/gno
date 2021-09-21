package gno

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

func (h Hashlet) Copy() Hashlet {
	return h
}

func (h Hashlet) Bytes() []byte {
	return h[:]
}

func (h Hashlet) IsZero() bool {
	return h == Hashlet{}
}

func HashBytes(bz []byte) (res Hashlet) {
	hash := sha256.Sum256(bz)
	copy(res[:], hash[:HashSize])
	return
}

func leafHash(bz []byte) (res Hashlet) {
	buf := make([]byte, 1+len(bz))
	buf[0] = 0x00
	copy(buf[1:], bz)
	res = HashBytes(buf)
	return
}

func innerHash(h1, h2 Hashlet) (res Hashlet) {
	buf := make([]byte, 1+HashSize*2)
	buf[0] = 0x01
	copy(buf[1:1+HashSize], h1[:])
	copy(buf[1+HashSize:], h2[:])
	res = HashBytes(buf)
	return
}

//----------------------------------------
// misc

func varintBytes(u int64) []byte {
	var buf [10]byte
	n := binary.PutVarint(buf[:], u)
	return buf[0:n]
}

func sizedBytes(bz []byte) []byte {
	bz2 := make([]byte, len(bz)+10)
	n := binary.PutVarint(bz2[:10], int64(len(bz)))
	copy(bz2[n:n+len(bz)], bz)
	return bz2[:n+len(bz)]
}

func isASCIIText(bz []byte) bool {
	if len(bz) == 0 {
		return false
	}
	for _, b := range bz {
		if 32 <= b && b <= 126 {
			// good
		} else {
			return false
		}
	}
	return true
}
