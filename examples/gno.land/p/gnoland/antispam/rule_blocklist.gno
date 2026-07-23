package antispam

import (
	"errors"
	"regexp"
	"strings"

	"gno.land/p/nt/avl"
)

var errPatternCapReached = errors.New("pattern cap reached (30 max)")

const (
	// blMaxPatterns caps the number of regex patterns.
	blMaxPatterns = 30
)

// Blocklist manages blocked addresses, allowed addresses, and regex patterns.
// Patterns are stored as strings for admin management (add/remove/iterate)
// and compiled into a single combined regex for matching. One regex scan
// per Score() call regardless of how many patterns are loaded.
type Blocklist struct {
	blocked  *avl.Tree      // address -> true
	allowed  *avl.Tree      // address -> true
	patterns *avl.Tree      // pattern string -> true (strings only, for management)
	combined *regexp.Regexp // single compiled (?:p1|p2|...|pN) for matching
}

// NewBlocklist creates an empty Blocklist.
func NewBlocklist() *Blocklist {
	return &Blocklist{
		blocked:  avl.NewTree(),
		allowed:  avl.NewTree(),
		patterns: avl.NewTree(),
	}
}

// Address blocking

// BlockAddress adds an address to the blocklist.
func (bl *Blocklist) BlockAddress(addr string) {
	bl.blocked.Set(addr, true)
}

// UnblockAddress removes an address from the blocklist.
func (bl *Blocklist) UnblockAddress(addr string) {
	bl.blocked.Remove(addr)
}

// IsBlocked returns true if the address is in the blocklist.
func (bl *Blocklist) IsBlocked(addr string) bool {
	return bl.blocked.Has(addr)
}

// BlockedCount returns the number of blocked addresses.
func (bl *Blocklist) BlockedCount() int {
	return bl.blocked.Size()
}

// Address allowlisting (whitelist)

// AllowAddress adds an address to the allowlist.
// Allowed addresses bypass all spam scoring.
func (bl *Blocklist) AllowAddress(addr string) {
	bl.allowed.Set(addr, true)
}

// RemoveAllow removes an address from the allowlist.
func (bl *Blocklist) RemoveAllow(addr string) {
	bl.allowed.Remove(addr)
}

// IsAllowed returns true if the address is in the allowlist.
func (bl *Blocklist) IsAllowed(addr string) bool {
	return bl.allowed.Has(addr)
}

// AllowedCount returns the number of allowed addresses.
func (bl *Blocklist) AllowedCount() int {
	return bl.allowed.Size()
}

// Pattern matching

// AddPattern validates and stores a regex pattern string.
// Returns an error if the pattern is invalid regex or the cap is reached.
// Rebuilds the single combined regex used for matching.
func (bl *Blocklist) AddPattern(pattern string) error {
	if bl.patterns.Size() >= blMaxPatterns && !bl.patterns.Has(pattern) {
		return errPatternCapReached
	}
	// Validate the pattern compiles before storing.
	if _, err := regexp.Compile(pattern); err != nil {
		return err
	}
	bl.patterns.Set(pattern, true)
	bl.rebuildCombined()
	return nil
}

// PatternCount returns the number of stored patterns.
func (bl *Blocklist) PatternCount() int {
	return bl.patterns.Size()
}

// IteratePatterns calls fn for each stored pattern string.
// Return true from fn to stop iteration early.
func (bl *Blocklist) IteratePatterns(fn func(pattern string) bool) {
	bl.patterns.Iterate("", "", func(key string, _ interface{}) bool {
		return fn(key)
	})
}

// RemovePattern removes a regex pattern and rebuilds the combined regex.
func (bl *Blocklist) RemovePattern(pattern string) {
	bl.patterns.Remove(pattern)
	bl.rebuildCombined()
}

// MatchesPattern checks content against the combined regex in a single scan.
// Returns true if any stored pattern matches.
func (bl *Blocklist) MatchesPattern(content string) bool {
	if content == "" || bl.combined == nil {
		return false
	}
	return bl.combined.MatchString(content)
}

// rebuildCombined compiles all patterns into a single alternation regex.
// Called when patterns are added or removed. Compilation cost is paid once
// at admin time, not per Score() call.
func (bl *Blocklist) rebuildCombined() {
	if bl.patterns.Size() == 0 {
		bl.combined = nil
		return
	}
	var b strings.Builder
	b.WriteString("(?:")
	first := true
	bl.patterns.Iterate("", "", func(key string, _ interface{}) bool {
		if !first {
			b.WriteByte('|')
		}
		b.WriteString(key)
		first = false
		return false
	})
	b.WriteByte(')')
	// Patterns were individually validated in AddPattern. Compilation of
	// (?:p1|p2|...) should always succeed. Panic on failure to surface
	// bugs immediately rather than silently disabling pattern matching.
	var err error
	bl.combined, err = regexp.Compile(b.String())
	if err != nil {
		panic("antispam: combined pattern compile failed: " + err.Error())
	}
}
