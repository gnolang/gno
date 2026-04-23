package amino

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// DumpWire returns a human-readable annotated trace of a raw amino/proto3
// wire-format byte stream. It is schema-less: it only knows the TLV
// structure (field number + Typ3 + value), not field names or types.
//
// Intended use is debug aid: when two codecs disagree on bytes, render
// both sides with DumpWire and diff the traces instead of hex dumps.
// See compareEncoding in tests/pb3_test.go for wiring into failure
// messages.
//
// The walker is best-effort: if it cannot decode a field key, it emits
// the remaining bytes as "<raw: ...>" and stops. ByteLength payloads
// are tentatively re-parsed as nested messages; if that fails they are
// emitted as bytes (hex for short, truncated with length for long).
//
// Output format (one logical event per line; nesting uses indent):
//
//	@0 field 1 Typ3ByteLength len=4
//	  @2 field 1 Typ3Varint = 42
//	@6 field 2 Typ3Varint = 7
//
// The leading @N is the byte offset into the outer slice, useful when
// cross-referencing with hex dumps.
func DumpWire(bz []byte) string {
	var sb strings.Builder
	dumpWireInto(&sb, bz, 0, "")
	return sb.String()
}

// dumpWireInto walks bz recursively, writing lines to sb. baseOff is
// the absolute offset of bz[0] into the outermost slice; indent is the
// prefix applied to each emitted line.
func dumpWireInto(sb *strings.Builder, bz []byte, baseOff int, indent string) {
	pos := 0
	for pos < len(bz) {
		abs := baseOff + pos
		fnum, typ3, n, err := DecodeFieldNumberAndTyp3(bz[pos:])
		if err != nil {
			fmt.Fprintf(sb, "%s@%d <bad field key: %v; remainder=%s>\n",
				indent, abs, err, shortHex(bz[pos:]))
			return
		}
		pos += n

		switch typ3 {
		case Typ3Varint:
			v, vn, verr := DecodeUvarint(bz[pos:])
			if verr != nil {
				fmt.Fprintf(sb, "%s@%d field %d Typ3Varint <bad: %v>\n",
					indent, abs, fnum, verr)
				return
			}
			// Print both unsigned and signed (zigzag-free) interpretations —
			// the reader picks the one that matches the schema they have
			// in mind. We don't know which is correct without a schema.
			fmt.Fprintf(sb, "%s@%d field %d Typ3Varint = %d (u=%d)\n",
				indent, abs, fnum, int64(v), v)
			pos += vn

		case Typ38Byte:
			if len(bz)-pos < 8 {
				fmt.Fprintf(sb, "%s@%d field %d Typ38Byte <truncated: have %d>\n",
					indent, abs, fnum, len(bz)-pos)
				return
			}
			fmt.Fprintf(sb, "%s@%d field %d Typ38Byte = %s\n",
				indent, abs, fnum, hex.EncodeToString(bz[pos:pos+8]))
			pos += 8

		case Typ34Byte:
			if len(bz)-pos < 4 {
				fmt.Fprintf(sb, "%s@%d field %d Typ34Byte <truncated: have %d>\n",
					indent, abs, fnum, len(bz)-pos)
				return
			}
			fmt.Fprintf(sb, "%s@%d field %d Typ34Byte = %s\n",
				indent, abs, fnum, hex.EncodeToString(bz[pos:pos+4]))
			pos += 4

		case Typ3ByteLength:
			payload, cn, perr := DecodeByteSlice(bz[pos:])
			if perr != nil {
				fmt.Fprintf(sb, "%s@%d field %d Typ3ByteLength <bad length: %v>\n",
					indent, abs, fnum, perr)
				return
			}
			payloadAbs := baseOff + pos
			// Skip the uvarint length prefix; payload begins at
			// payloadAbs + (cn - len(payload)).
			payloadStart := payloadAbs + (cn - len(payload))

			if len(payload) == 0 {
				// Empty ByteLength — legitimately either an empty message,
				// an empty string, or in pointer-lists with amino:"nil_elements"
				// a positional-nil sentinel. Annotate so readers notice.
				fmt.Fprintf(sb, "%s@%d field %d Typ3ByteLength len=0 (empty / possible nil sentinel)\n",
					indent, abs, fnum)
			} else if couldBeMessage(payload) {
				fmt.Fprintf(sb, "%s@%d field %d Typ3ByteLength len=%d\n",
					indent, abs, fnum, len(payload))
				dumpWireInto(sb, payload, payloadStart, indent+"  ")
			} else {
				fmt.Fprintf(sb, "%s@%d field %d Typ3ByteLength len=%d bytes=%s\n",
					indent, abs, fnum, len(payload), shortHex(payload))
			}
			pos += cn

		default:
			fmt.Fprintf(sb, "%s@%d field %d <unknown Typ3 %d>; remainder=%s\n",
				indent, abs, fnum, typ3, shortHex(bz[pos:]))
			return
		}
	}
}

// couldBeMessage heuristically decides whether to try recursive parsing.
// We parse the whole payload in dry-run: if every field key decodes and
// every length prefix fits, treat it as a nested message. This matches
// the way the consensus code itself navigates nested types.
//
// Short payloads (<2 bytes) or those with bytes that can't start a
// valid field key are classified as raw.
func couldBeMessage(bz []byte) bool {
	if len(bz) < 2 {
		return false
	}
	pos := 0
	for pos < len(bz) {
		fnum, typ3, n, err := DecodeFieldNumberAndTyp3(bz[pos:])
		if err != nil || fnum == 0 {
			return false
		}
		pos += n
		switch typ3 {
		case Typ3Varint:
			_, vn, verr := DecodeUvarint(bz[pos:])
			if verr != nil {
				return false
			}
			pos += vn
		case Typ38Byte:
			if len(bz)-pos < 8 {
				return false
			}
			pos += 8
		case Typ34Byte:
			if len(bz)-pos < 4 {
				return false
			}
			pos += 4
		case Typ3ByteLength:
			payload, cn, perr := DecodeByteSlice(bz[pos:])
			if perr != nil {
				return false
			}
			_ = payload
			pos += cn
		default:
			return false
		}
	}
	return pos == len(bz)
}

// shortHex returns a hex dump, truncated after 32 bytes to keep the
// trace readable. Long payloads get "...(Nmore)" suffix.
func shortHex(bz []byte) string {
	const cap = 32
	if len(bz) <= cap {
		return hex.EncodeToString(bz)
	}
	return fmt.Sprintf("%s...(%dmore)", hex.EncodeToString(bz[:cap]), len(bz)-cap)
}
