> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `cford32` - Crockford Base32 encoding

Modified base32 encoding using the [Crockford alphabet](https://www.crockford.com/base32.html). Designed to be human-readable, error-resistant, and pronounceable: the ambiguous characters `I`, `L`, `O`, `U` are excluded from the encoding, and decoding accepts `I`/`L` as `1` and `O` as `0`. Output is never padded.

## Usage

```go
import "gno.land/p/nt/cford32/v0"

// Byte slice encode/decode.
encoded := cford32.EncodeToString([]byte("hello"))  // uppercase, no padding
decoded, err := cford32.DecodeString(encoded)        // []byte("hello")

// Lowercase variant.
lower := cford32.EncodeToStringLower([]byte("hello"))

// Compact uint64 encoding: 7 bytes for id < 2^34, else 13 bytes.
enc := cford32.PutCompact(42)
back, _ := cford32.Uint64(enc) // 42

// Full fixed-width uint64 encoding (always 13 bytes).
full := cford32.PutUint64(42)
```

## API

```go
// Errors.
type CorruptInputError int64
func (e CorruptInputError) Error() string

// Length helpers.
func DecodedLen(n int) int
func EncodedLen(n int) int

// Byte slice encoding.
func Encode(dst, src []byte)                          // uppercase
func EncodeLower(dst, src []byte)                     // lowercase
func EncodeToString(src []byte) string                // uppercase
func EncodeToStringLower(src []byte) string           // lowercase
func AppendEncode(dst, src []byte) []byte
func AppendEncodeLower(dst, src []byte) []byte

// Byte slice decoding. Case-insensitive; ignores \r and \n.
func Decode(dst, src []byte) (n int, err error)
func DecodeString(s string) ([]byte, error)
func AppendDecode(dst, src []byte) ([]byte, error)

// uint64 encoding.
func PutUint64(id uint64) [13]byte                    // full, uppercase
func PutUint64Lower(id uint64) [13]byte               // full, lowercase
func PutCompact(id uint64) []byte                     // 7 bytes if id < 2^34, else 13, lowercase
func AppendCompact(id uint64, b []byte) []byte
func Uint64(b []byte) (uint64, error)                 // accepts both compact (7) and full (13)

// Streaming I/O.
func NewEncoder(w io.Writer) io.WriteCloser
func NewEncoderLower(w io.Writer) io.WriteCloser
func NewDecoder(r io.Reader) io.Reader
```

## Notes

- Alphabet: `0123456789ABCDEFGHJKMNPQRSTVWXYZ` (no `I`, `L`, `O`, `U`).
- Decoding is case-insensitive; `I`/`i`/`L`/`l` decode as `1`, and `O`/`o` decode as `0`.
- The compact uint64 encoding preserves lexicographic order with numeric order, making encoded IDs suitable as ordered keys.
- The compact and full uint64 encodings are unambiguously distinguished by their first character: `0`-`f` indicates compact (7 bytes), `g`-`z` indicates full (13 bytes).
- Values in `[0, 2^34)` have BOTH a compact and a full encoding. Pick one scheme per key space and stick to it: mixing both for the same value breaks the lexicographic-order property. `PutCompact` rolls over from compact to full at `2^34` automatically, which is safe as long as everything in that space is generated the same way.
- For sequential IDs, see [`gno.land/p/nt/seqid/v0`](../../seqid/v0).
