package state

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	OIDPattern  = regexp.MustCompile(`^[A-Fa-f0-9]{40}:\d+$`)
	FilePattern = regexp.MustCompile(`^[A-Za-z0-9_./-]+\.gno$`)
)

const (
	MaxStateIDLength    = 256
	MaxFragmentLine     = 1_000_000
	MaxFragmentFileSize = 256 * 1024
)

var (
	ErrInvalidOID    = errors.New("invalid object id")
	ErrInvalidTID    = errors.New("invalid type id")
	ErrInvalidFile   = errors.New("invalid file")
	ErrInvalidHeight = errors.New("invalid height")
	ErrInvalidLine   = errors.New("invalid line")
	ErrInvalidOffset = errors.New("invalid offset")
	ErrInvalidLimit  = errors.New("invalid limit")
)

func ValidateOID(s string) error {
	if len(s) > MaxStateIDLength || !OIDPattern.MatchString(s) {
		return ErrInvalidOID
	}
	return nil
}

// ValidateTID bounds the attacker-controlled &tid= param. A Gno TypeID
// is a human-readable string (e.g. "gno.land/r/demo/foo.Bar", "int"),
// not a hash — so cap length and reject control chars, nothing more.
func ValidateTID(s string) error {
	if s == "" || len(s) > MaxStateIDLength || strings.ContainsFunc(s, unicode.IsControl) {
		return ErrInvalidTID
	}
	return nil
}

func ValidateFile(s string) error {
	if len(s) > MaxStateIDLength || !FilePattern.MatchString(s) {
		return ErrInvalidFile
	}
	// FilePattern's char class accepts `.` and `/`, which composes into
	// `..` and `/..` traversal segments. The RPC fetcher path-joins onto
	// the pkgPath and cleans the result, so a traversal would resolve to
	// a different on-chain package (cache pollution). Reject explicitly.
	if s == ".." || strings.HasPrefix(s, "../") || strings.HasSuffix(s, "/..") ||
		strings.Contains(s, "/../") || strings.HasPrefix(s, "/") {
		return ErrInvalidFile
	}
	return nil
}

// ValidateHeight returns 0 for empty input (meaning "latest").
func ValidateHeight(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n < 0 {
		return 0, ErrInvalidHeight
	}
	return n, nil
}

func ValidateLine(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > MaxFragmentLine {
		return 0, ErrInvalidLine
	}
	return n, nil
}

// ValidateOffset bounds the attacker-controlled `offset` pagination param.
// Empty input → 0 (first page). Anything else must parse to a non-negative
// integer; oversize values survive validation because the page handler
// clamps offset to total after the decode (out-of-range pages render empty).
func ValidateOffset(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, ErrInvalidOffset
	}
	return n, nil
}

// ValidateLimit bounds the attacker-controlled `limit` pagination param.
// Empty input → maxTopLevelDecls (default page size). Anything else must
// parse to a positive integer; values above the cap silently clamp so the
// per-page fragment fan-out budget always holds.
func ValidateLimit(s string) (int, error) {
	if s == "" {
		return maxTopLevelDecls, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, ErrInvalidLimit
	}
	if n > maxTopLevelDecls {
		n = maxTopLevelDecls
	}
	return n, nil
}

// View mode constants for the ?state&view= query param. Used everywhere
// the Go side compares against the mode to avoid string-literal sprawl.
// Templates still use raw literals — that's UI presentation, not config.
const (
	ViewModePretty = "pretty"
	ViewModeTree   = "tree"
)

// CanonicalViewMode normalizes the ?state&view=… query param.
// URL-driven so the nginx cache key stays URL-only (no Vary: Cookie split).
func CanonicalViewMode(s string) string {
	if s == ViewModeTree {
		return ViewModeTree
	}
	return ViewModePretty
}
