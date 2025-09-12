# cford32

```
package cford32 // import "gno.land/p/nt/cford32"

Package cford32 implements a base32-like encoding/decoding package, with the
encoding scheme specified by Douglas Crockford.

From the website, the requirements of said encoding scheme are to:

  - Be human readable and machine readable.
  - Be compact. Humans have difficulty in manipulating long strings of arbitrary
    symbols.
  - Be error resistant. Entering the symbols must not require keyboarding
    gymnastics.
  - Be pronounceable. Humans should be able to accurately transmit the symbols
    to other humans using a telephone.

This is slightly different from a simple difference in encoding table from
the Go's stdlib `encoding/base32`, as when decoding the characters i I l L are
parsed as 1, and o O is parsed as 0.

This package additionally provides ways to encode uint64's efficiently, as well
as efficient encoding to a lowercase variation of the encoding. The encodings
never use paddings.

# Uint64 Encoding

Aside from lower/uppercase encoding, there is a compact encoding, allowing to
encode all values in [0,2^34), and the full encoding, allowing all values in
[0,2^64). The compact encoding uses 7 characters, and the full encoding uses 13
characters. Both are parsed unambiguously by the Uint64 decoder.

The compact encodings have the first character between ['0','f'], while the
full encoding's first character ranges between ['g','z']. Practically, in your
usage of the package, you should consider which one to use and stick with it,
while considering that the compact encoding, once it reaches 2^34, automatically
switches to the full encoding. The properties of the generated strings are still
maintained: for instance, any two encoded uint64s x,y consistently generated
with the compact encoding, if the numeric value is x < y, will also be x < y in
lexical ordering. However, values [0,2^34) have a "double encoding", which if
mixed together lose the lexical ordering property.

The Uint64 encoding is most useful for generating string versions of Uint64 IDs.
Practically, it allows you to retain sleek and compact IDs for your application
for the first 2^34 (>17 billion) entities, while seamlessly rolling over to the
full encoding should you exceed that. You are encouraged to use it unless you
have a requirement or preferences for IDs consistently being always the same
size.

To use the cford32 encoding for IDs, you may want to consider using package
gno.land/p/nt/seqid.

[specified by Douglas Crockford]: https://www.crockford.com/base32.html

func AppendCompact(id uint64, b []byte) []byte
func AppendDecode(dst, src []byte) ([]byte, error)
func AppendEncode(dst, src []byte) []byte
func AppendEncodeLower(dst, src []byte) []byte
func Decode(dst, src []byte) (n int, err error)
func DecodeString(s string) ([]byte, error)
func DecodedLen(n int) int
func Encode(dst, src []byte)
func EncodeLower(dst, src []byte)
func EncodeToString(src []byte) string
func EncodeToStringLower(src []byte) string
func EncodedLen(n int) int
func NewDecoder(r io.Reader) io.Reader
func NewEncoder(w io.Writer) io.WriteCloser
func NewEncoderLower(w io.Writer) io.WriteCloser
func PutCompact(id uint64) []byte
func PutUint64(id uint64) [13]byte
func PutUint64Lower(id uint64) [13]byte
func Uint64(b []byte) (uint64, error)
type CorruptInputError int64
```
