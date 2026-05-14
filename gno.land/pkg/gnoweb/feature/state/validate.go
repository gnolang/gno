package state

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Validation patterns per ADR-004 §2.
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
	// a different on-chain package (cache pollution + ADR-004 §Threat
	// model claim violation). Reject explicitly to keep the contract.
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

// CanonicalViewMode normalizes the ?state&view=… query param. Returns
// "pretty" (default) for empty or unknown input, "tree" for "tree".
// URL-driven view-mode (ADR-004 §6 Option B) keeps the nginx cache key
// URL-only and avoids the SSR cookie + Vary: Cookie split.
func CanonicalViewMode(s string) string {
	if s == "tree" {
		return "tree"
	}
	return "pretty"
}
