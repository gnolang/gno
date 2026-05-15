package state

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

// encodeInt64LE renders a little-endian int64 as base64, matching how
// Amino emits PrimitiveType("32") fields in qobject_json responses.
func encodeInt64LE(v int64) string {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	return base64.StdEncoding.EncodeToString(b[:])
}

// previewStructBody returns a minimal qobject_json shape for a 2-field
// struct, sufficient to seed StateObject fixtures in page/bench tests.
func previewStructBody(oid string, val0, val1 int) []byte {
	return []byte(fmt.Sprintf(`{
		"objectid": %q,
		"value": {
			"@type": "/gno.StructValue",
			"Fields": [
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"},
				{"T": {"@type": "/gno.PrimitiveType", "value": "32"}, "N": "%s"}
			]
		}
	}`, oid, encodeInt64LE(int64(val0)), encodeInt64LE(int64(val1))))
}
