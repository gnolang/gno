> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `seqid` - Sequential IDs

Sequential ID generator producing ordered binary and string representations suitable for use as AVL tree keys. String IDs use [cford32](../../cford32/v0)'s compact encoding and preserve lexicographic ordering.

## Usage

```go
import (
    "gno.land/p/nt/avl/v0"
    "gno.land/p/nt/seqid/v0"
)

var (
    id    seqid.ID
    users avl.Tree
)

func NewUser(name string) {
    user := &User{Name: name}

    // String() is human-friendly and preserves ordering.
    users.Set(id.Next().String(), user)

    // Or persist the binary form as a fixed-width 8-byte AVL key.
    users.Set(id.Next().Binary(), user)
}

// Recover an ID from user input (case-insensitive, sanitized).
func Lookup(raw string) (seqid.ID, error) {
    return seqid.FromString(raw)
}
```

## API

```go
// An ID is a sequential ID. The zero value is valid; the first
// Next() call returns 1.
type ID uint64

// Next advances the ID and returns the new value. Panics on overflow.
func (i *ID) Next() ID

// TryNext is like Next but returns false instead of panicking on overflow.
func (i *ID) TryNext() (ID, bool)

// Binary returns a fixed 8-byte big-endian encoding of the ID, suitable
// as an AVL key. Lexicographic order matches numeric order.
func (i ID) Binary() string

// String returns the cford32 compact encoding of the ID: 7 bytes for
// IDs in [0, 2^34), 13 bytes after that. Lexicographic order matches
// numeric order across the rollover.
func (i ID) String() string

// FromBinary parses a value produced by Binary.
func FromBinary(b string) (ID, bool)

// FromString parses a cford32-encoded ID. Case-insensitive; maps
// I/L to 1 and O to 0. Always re-encode user input via FromString
// then String() before using it as a key.
func FromString(b string) (ID, error)
```

## Notes

- `Binary()` is the cheapest and most compact key (8 bytes, fixed width). Prefer it for internal storage. The keys work with any `ITree` (`gno.land/p/nt/avl/v0` or `gno.land/p/nt/bptree/v0`); their monotonic order suits bptree's append path especially well.
- `String()` is human-friendly and URL-safe; use it for IDs surfaced to users.
- Because cford32 accepts multiple spellings for the same value, always normalize external input through `FromString` then `String()` before using it as a lookup key.
