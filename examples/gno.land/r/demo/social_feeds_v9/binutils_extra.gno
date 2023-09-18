package social_feeds

import (
	"encoding/binary"
)

func EncodeLengthPrefixedStringUint32BE(s string) []byte {
	b := make([]byte, 4+len(s))
	binary.BigEndian.PutUint32(b, uint32(len(s)))
	copy(b[4:], s)
	return b
}
