> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `bylaws` - Governing documents amended by verifiable patches

Stores a DAO's governing documents — bylaws and mandates — as named plaintext files, and amends them with diff patches that are pinned to the text they were written against.

Documents are keyed by a slash-separated path such as `mandates/treasury.md`. Folders are a naming convention over the path, not stored objects: a folder exists exactly when some document path has it as a prefix.

The only mutation is `Apply`. A `Patch` carries the sha256 of the document text it was diffed against plus the edit script that turns that text into the proposed one, so `Apply` rejects the patch if the document changed in the meantime — optimistic concurrency, no silent clobbering. An amendment whose result is empty removes the document, so a stored document is never empty.

The package is **governance-agnostic**: it decides nothing about *who* may amend. A consuming realm gates `Apply` behind its own vote and keeps the `*Bylaws` handle private.

## Usage

```go
package mydao

import "gno.land/p/nt/bylaws/v0"

var docs = bylaws.New() // keep this handle private: Apply mutates

// A proposer builds a patch off-chain or in a query, then submits the
// encoded string as a proposal argument.
func ProposeAmendment(path, proposed string) (string, error) {
    p, err := docs.Diff(path, proposed)
    if err != nil {
        return "", err
    }
    return p.Encode(), nil
}

// The DAO calls this only after its own vote passes.
func executeAmendment(encoded string) error {
    p, err := bylaws.DecodePatch(encoded)
    if err != nil {
        return err
    }
    return docs.Apply(p) // fails if the document changed since the diff
}

func Render(path string) string {
    text, ok := docs.Get(path)
    if !ok {
        return "not found"
    }
    return text
}
```

Show voters what a pending patch does before they vote on it:

```go
base, _ := docs.Get(p.Path)
summary, err := p.Format(base) // "= 12 unchanged", "- …", "+ …"
```

## API

```go
type Bylaws struct{ /* unexported */ }

func New() *Bylaws

// Read
func (b *Bylaws) Get(path string) (string, bool)
func (b *Bylaws) Has(path string) bool
func (b *Bylaws) Size() int
func (b *Bylaws) Hash(path string) string // hex sha256, "" if absent
func (b *Bylaws) List(prefix string) []string
func (b *Bylaws) Iterate(prefix string, fn func(path, text string) bool) bool

// Write — the only mutation
func (b *Bylaws) Apply(p Patch) error

// Patches
func (b *Bylaws) Diff(path, proposed string) (Patch, error)
func DiffTexts(path, base, proposed string, exists bool) (Patch, error)
func DecodePatch(s string) (Patch, error)

type Patch struct {
    Path string // target document path
    Base string // hex sha256 of the base text; "" pins "document absent"
    Ops  []Op   // edit script, replayed in order
}

func (p Patch) Encode() string
func (p Patch) Format(base string) (string, error)
func (p Patch) IsCreate() bool
func (p Patch) IsRemove() bool
func (p Patch) IsNoop() bool

type Op struct {
    Type OpType // OpKeep, OpDelete, OpInsert
    N    int    // rune count (Keep and Delete)
    Text string // inserted literal (Insert)
}

func HashText(text string) string
func IsValidPath(path string) bool

const (
    MaxPathLen = 200
    MaxDocLen  = 64 * 1024
    MaxOps     = 16 * 1024
)

var (
    ErrInvalidPath  // path is not a valid document path
    ErrInvalidPatch // patch is malformed, or its script does not fit the base
    ErrInvalidText  // text is not valid UTF-8
    ErrStalePatch   // the document changed since the patch base
    ErrDocTooLarge  // result exceeds MaxDocLen
)
```

A valid path is one or more non-empty `/`-separated segments of `[a-zA-Z0-9._-]`, with no segment being `.` or `..`. The restricted charset keeps paths render- and link-safe and the patch encoding delimiter-free.

`List` and `Iterate` take a **raw path prefix**, not a folder name: pass `"mandates/"` with the trailing slash to scope to a folder, since `"mandates"` also matches a sibling file like `mandates-old.md`.

## Patch encoding

`Encode` produces a single compact string fit for a transaction argument:

```
v0:<path>:<base>:K12;D5;I8:new text;
```

`K<n>` and `D<n>` carry rune counts and `I<len>:<bytes>` carries a length-prefixed literal, so no escaping is needed. Only inserts carry bytes — a small edit to a large document stays a small patch. `DecodePatch` is the exact inverse and accepts only canonical output; it validates shape, not whether the script fits a real document, which is `Apply`'s job.

## Notes

- **Keep the `*Bylaws` handle private.** `Apply` mutates, so exposing the handle to untrusted callers hands them amendment rights under your realm's authority. Expose reads through your own wrapper functions.
- **`Apply` is all-or-nothing.** On any error the document set is unchanged: the base hash must match, and the edit script must consume the base text exactly, so a script that does not fit fails rather than producing garbage.
- **Diff falls back to a whole-region replacement on large edits.** Myers diff is `O((N+M)·D)` in memory, so past a rune budget `DiffTexts` replaces the changed middle in one delete+insert. Common prefix and suffix are trimmed first, so ordinary human edits — small changes to a large document — stay minimal.
- **`Format` output is raw text and insertions are proposer-controlled.** A realm rendering a patch summary as markdown must escape it; use `gno.land/p/nt/markdown/sanitize/v0`. The formatter marks whole multi-line literals so an insertion containing `"\n- fake"` cannot masquerade as the summary's own markers, but that is not a substitute for escaping.
- **`IsNoop` is shape-based.** It detects a script that only keeps text; a hand-built script that deletes and reinserts identical text is not detected. `Diff` never produces one.
- A patch with an empty `Base` pins "this document must not exist", i.e. it is a create. A patch whose script deletes everything and inserts nothing removes the document.
