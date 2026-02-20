package antispam

import (
	"errors"
	"regexp"

	"gno.land/p/nt/avl"
)

var errPatternCapReached = errors.New("pattern cap reached (30 max)")

const (
	// blMaxPatterns caps the number of regex patterns to prevent O(P*n)
	// cost growth. Each pattern runs a full regex match against content.
	blMaxPatterns = 30
)

// Blocklist manages blocked addresses, allowed addresses, and regex patterns.
// It uses AVL trees for deterministic iteration.
type Blocklist struct {
	blocked  *avl.Tree // address -> true
	allowed  *avl.Tree // address -> true
	patterns *avl.Tree // pattern string -> *regexp.Regexp
}

// NewBlocklist creates an empty Blocklist.
func NewBlocklist() *Blocklist {
	return &Blocklist{
		blocked:  avl.NewTree(),
		allowed:  avl.NewTree(),
		patterns: avl.NewTree(),
	}
}

// ---- Address blocking

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

// ---- Address allowlisting (whitelist)

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

// ---- Pattern matching

// AddPattern compiles and stores a regex pattern.
// Returns an error if the pattern is invalid or the cap is reached.
func (bl *Blocklist) AddPattern(pattern string) error {
	if bl.patterns.Size() >= blMaxPatterns && !bl.patterns.Has(pattern) {
		return errPatternCapReached
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	bl.patterns.Set(pattern, re)
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

// RemovePattern removes a regex pattern.
func (bl *Blocklist) RemovePattern(pattern string) {
	bl.patterns.Remove(pattern)
}

// MatchesPattern checks content against all stored regex patterns.
// Returns whether any pattern matched and the first matching pattern.
func (bl *Blocklist) MatchesPattern(content string) (bool, string) {
	if content == "" {
		return false, ""
	}
	matched := false
	matchedPattern := ""
	bl.patterns.Iterate("", "", func(key string, value interface{}) bool {
		re := value.(*regexp.Regexp)
		if re.MatchString(content) {
			matched = true
			matchedPattern = key
			return true // stop iteration
		}
		return false
	})
	return matched, matchedPattern
}

// ---- Scoring

// ScoreBlocklist evaluates author and content against the blocklist.
// Allowed addresses return 0 immediately.
// Blocked addresses return WeightBlockedAddress.
// Pattern matches return WeightBlockedPattern.
func ScoreBlocklist(author, content string, bl *Blocklist) (int, string) {
	if bl == nil {
		return 0, ""
	}

	// Allowlist bypasses everything
	if bl.IsAllowed(author) {
		return 0, ""
	}

	// Blocked address
	if bl.IsBlocked(author) {
		return WeightBlockedAddress, "BLOCKED_ADDRESS"
	}

	// Pattern matching
	if matched, _ := bl.MatchesPattern(content); matched {
		return WeightBlockedPattern, "BLOCKED_PATTERN"
	}

	return 0, ""
}
