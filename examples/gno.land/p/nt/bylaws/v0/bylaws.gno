// Package bylaws stores a DAO's governing documents — bylaws and
// mandates — as named plaintext files and amends them with verifiable
// diff patches.
//
// Documents are keyed by a slash-separated path ("mandates/treasury.md");
// folders are a naming convention over the path, not stored objects — a
// folder exists exactly when a document path has it as a prefix. The only
// mutation is Apply: a Patch carries the sha256 of the document text it
// was diffed against plus the edit script that transforms that text into
// the proposed one. Apply rejects the patch when the document has changed
// since (optimistic concurrency, no clobbering) and otherwise replays the
// script. An amendment whose result is empty removes the document, so a
// stored document is never empty.
//
// The package is governance-agnostic: it decides nothing about WHO may
// amend. A consuming realm (e.g. a DAO) gates Apply behind its own vote
// and keeps the *Bylaws handle private — Apply mutates, so the handle
// must never be exposed to untrusted callers.
package bylaws

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"gno.land/p/nt/avl/v0"
)

const (
	// MaxPathLen bounds a document path's byte length.
	MaxPathLen = 200

	// MaxDocLen bounds a document's byte length. Bylaws are human-written
	// prose; the cap keeps documents renderable and patch replay bounded.
	MaxDocLen = 64 * 1024
)

var (
	ErrInvalidPath  = errors.New("bylaws: invalid document path")
	ErrInvalidPatch = errors.New("bylaws: invalid patch")
	ErrInvalidText  = errors.New("bylaws: text is not valid UTF-8")
	ErrStalePatch   = errors.New("bylaws: document changed since the patch base")
	ErrDocTooLarge  = errors.New("bylaws: document exceeds the maximum size")
)

// Bylaws is one DAO's set of governing documents, keyed by path.
type Bylaws struct {
	docs *avl.Tree // path (string) -> document text (string, never empty)
}

// New creates an empty document set.
func New() *Bylaws {
	return &Bylaws{docs: avl.NewTree()}
}

// Get returns a document's text and whether it exists.
func (b *Bylaws) Get(path string) (string, bool) {
	if v := b.docs.Get(path); v != nil {
		return v.(string), true
	}
	return "", false
}

// Has reports whether a document exists.
func (b *Bylaws) Has(path string) bool {
	return b.docs.Has(path)
}

// Size returns the number of documents.
func (b *Bylaws) Size() int {
	return b.docs.Size()
}

// Hash returns the hex sha256 of a document's text, or an empty string
// when the document does not exist. It is the base a Patch must pin to
// amend the document (an empty hash pins "the document must not exist").
func (b *Bylaws) Hash(path string) string {
	if text, ok := b.Get(path); ok {
		return HashText(text)
	}
	return ""
}

// List returns the sorted document paths under a prefix. An empty prefix
// lists every document. The prefix is a raw path prefix: include the
// trailing slash to scope to a folder (e.g. "mandates/"), or "mandates"
// also matches a sibling file like "mandates-old.md".
func (b *Bylaws) List(prefix string) []string {
	paths := []string{}
	b.Iterate(prefix, func(path, _ string) bool {
		paths = append(paths, path)
		return false
	})
	return paths
}

// Iterate walks the documents under a prefix in sorted path order until
// fn returns true. It returns true when the walk was stopped by fn.
func (b *Bylaws) Iterate(prefix string, fn func(path, text string) bool) bool {
	end := ""
	if prefix != "" {
		// Path bytes are all < 0x7f (see IsValidPath), so every key with
		// the prefix sorts before prefix+"\x7f".
		end = prefix + "\x7f"
	}
	return b.docs.Iterate(prefix, end, func(key string, value any) bool {
		return fn(key, value.(string))
	})
}

// HashText returns the hex sha256 of a text.
func HashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

// IsValidPath reports whether a path names a document: one or more
// non-empty "/"-separated segments of [a-zA-Z0-9._-] characters, where no
// segment is "." or "..". The restricted charset keeps paths render- and
// link-safe and the patch encoding delimiter-free.
func IsValidPath(path string) bool {
	if path == "" || len(path) > MaxPathLen {
		return false
	}
	segStart := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			seg := path[segStart:i]
			if seg == "" || seg == "." || seg == ".." {
				return false
			}
			segStart = i + 1
			continue
		}
		if !isPathChar(path[i]) {
			return false
		}
	}
	return true
}

func isPathChar(c byte) bool {
	return c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z' ||
		c >= '0' && c <= '9' ||
		c == '.' || c == '_' || c == '-'
}
