package antisquatting

import (
	"regexp"
	"strings"
	"time"
	
	"gno.land/p/demo/avl"
)

// DefaultNameClassifier implements NameClassifier with configurable rules
type DefaultNameClassifier struct {
	// Configuration
	shortNameThreshold   int           // Names shorter than this are high-value
	minimumBid          int64         // Base minimum bid
	shortNameMultiplier int64         // Multiplier for short names
	systemNameMultiplier int64        // Multiplier for system names
	
	// Durations
	defaultCommitDuration time.Duration
	defaultRevealDuration time.Duration
	shortNameCommitDuration time.Duration
	systemNameCommitDuration time.Duration
	
	// Deposits
	defaultDeposit   int64
	shortNameDeposit int64
	systemNameDeposit int64
	
	// Reserved words and patterns
	systemWords     *avl.Tree // string -> bool
	trademarkWords  *avl.Tree // string -> bool
	commonWords     *avl.Tree // string -> bool
	
	// Regex patterns
	alphanumericOnly *regexp.Regexp
}

// NewDefaultNameClassifier creates a new classifier with default settings
func NewDefaultNameClassifier() *DefaultNameClassifier {
	classifier := &DefaultNameClassifier{
		// Basic configuration
		shortNameThreshold:   4,
		minimumBid:          1_000_000, // 1 GNOT
		shortNameMultiplier: 10,
		systemNameMultiplier: 50,
		
		// Durations
		defaultCommitDuration:    24 * time.Hour,      // 1 day
		defaultRevealDuration:    24 * time.Hour,      // 1 day
		shortNameCommitDuration:  7 * 24 * time.Hour,  // 1 week
		systemNameCommitDuration: 7 * 24 * time.Hour,  // 1 week
		
		// Deposits
		defaultDeposit:    100_000, // 0.1 GNOT
		shortNameDeposit:  1_000_000, // 1 GNOT
		systemNameDeposit: 5_000_000, // 5 GNOT
		
		// Initialize trees
		systemWords:    avl.NewTree(),
		trademarkWords: avl.NewTree(),
		commonWords:    avl.NewTree(),
		
		// Compile regex
		alphanumericOnly: regexp.MustCompile("^[a-zA-Z0-9_]+$"),
	}
	
	// Initialize system reserved words
	classifier.initializeSystemWords()
	classifier.initializeCommonWords()
	
	return classifier
}

// IsHighValue determines if a name requires auction-based registration
func (c *DefaultNameClassifier) IsHighValue(name string) bool {
	// Normalize name
	name = strings.ToLower(strings.TrimSpace(name))
	
	// Empty or invalid names are not high-value
	if name == "" || !c.alphanumericOnly.MatchString(name) {
		return false
	}
	
	// Short names are always high-value
	if len(name) < c.shortNameThreshold {
		return true
	}
	
	// System reserved words
	if c.isSystemWord(name) {
		return true
	}
	
	// Trademark protected words
	if c.isTrademarkWord(name) {
		return true
	}
	
	// Common dictionary words
	if c.isCommonWord(name) {
		return true
	}
	
	// Check for patterns that indicate high value
	if c.hasHighValuePattern(name) {
		return true
	}
	
	return false
}

// GetMinimumBid returns the minimum bid for a name
func (c *DefaultNameClassifier) GetMinimumBid(name string) int64 {
	name = strings.ToLower(strings.TrimSpace(name))
	
	// Base minimum bid
	minBid := c.minimumBid
	
	// Apply multipliers based on name characteristics
	if len(name) < c.shortNameThreshold {
		minBid *= c.shortNameMultiplier
	}
	
	if c.isSystemWord(name) {
		minBid *= c.systemNameMultiplier
	}
	
	// Additional multipliers for very short names
	if len(name) == 1 {
		minBid *= 100
	} else if len(name) == 2 {
		minBid *= 50
	} else if len(name) == 3 {
		minBid *= 20
	}
	
	return minBid
}

// GetAuctionDuration returns the commit phase duration
func (c *DefaultNameClassifier) GetAuctionDuration(name string) time.Duration {
	name = strings.ToLower(strings.TrimSpace(name))
	
	// System names get longer auction periods
	if c.isSystemWord(name) {
		return c.systemNameCommitDuration
	}
	
	// Short names get longer auction periods
	if len(name) < c.shortNameThreshold {
		return c.shortNameCommitDuration
	}
	
	return c.defaultCommitDuration
}

// GetRevealDuration returns the reveal phase duration
func (c *DefaultNameClassifier) GetRevealDuration(name string) time.Duration {
	// Reveal duration is typically shorter and more consistent
	return c.defaultRevealDuration
}

// GetMinimumDeposit returns the minimum deposit required
func (c *DefaultNameClassifier) GetMinimumDeposit(name string) int64 {
	name = strings.ToLower(strings.TrimSpace(name))
	
	// System names require higher deposits
	if c.isSystemWord(name) {
		return c.systemNameDeposit
	}
	
	// Short names require higher deposits
	if len(name) < c.shortNameThreshold {
		return c.shortNameDeposit
	}
	
	return c.defaultDeposit
}

// Helper methods

func (c *DefaultNameClassifier) isSystemWord(name string) bool {
	_, exists := c.systemWords.Get(name)
	return exists
}

func (c *DefaultNameClassifier) isTrademarkWord(name string) bool {
	_, exists := c.trademarkWords.Get(name)
	return exists
}

func (c *DefaultNameClassifier) isCommonWord(name string) bool {
	_, exists := c.commonWords.Get(name)
	return exists
}

func (c *DefaultNameClassifier) hasHighValuePattern(name string) bool {
	// Check for patterns like repeated characters, sequences, etc.
	
	// All same character (aaa, 111, etc.)
	if c.isRepeatedChar(name) {
		return true
	}
	
	// Sequential characters (abc, 123, etc.)
	if c.isSequential(name) {
		return true
	}
	
	// Palindromes of reasonable length
	if len(name) >= 3 && len(name) <= 7 && c.isPalindrome(name) {
		return true
	}
	
	return false
}

func (c *DefaultNameClassifier) isRepeatedChar(name string) bool {
	if len(name) < 3 {
		return false
	}
	
	first := name[0]
	for i := 1; i < len(name); i++ {
		if name[i] != first {
			return false
		}
	}
	return true
}

func (c *DefaultNameClassifier) isSequential(name string) bool {
	if len(name) < 3 {
		return false
	}
	
	// Check ascending sequence
	ascending := true
	for i := 1; i < len(name); i++ {
		if name[i] != name[i-1]+1 {
			ascending = false
			break
		}
	}
	
	if ascending {
		return true
	}
	
	// Check descending sequence
	for i := 1; i < len(name); i++ {
		if name[i] != name[i-1]-1 {
			return false
		}
	}
	
	return true
}

func (c *DefaultNameClassifier) isPalindrome(name string) bool {
	for i := 0; i < len(name)/2; i++ {
		if name[i] != name[len(name)-1-i] {
			return false
		}
	}
	return true
}

func (c *DefaultNameClassifier) initializeSystemWords() {
	systemWords := []string{
		"admin", "administrator", "root", "system", "sys", "kernel",
		"gov", "government", "governance", "dao", "treasury",
		"api", "www", "mail", "email", "ftp", "ssh", "dns",
		"test", "testing", "dev", "development", "prod", "production",
		"staging", "demo", "example", "sample", "null", "undefined",
		"true", "false", "yes", "no", "on", "off", "enable", "disable",
		"public", "private", "secret", "hidden", "internal", "external",
		"user", "users", "member", "members", "guest", "anonymous",
		"support", "help", "info", "about", "contact", "feedback",
		"security", "auth", "login", "logout", "register", "signup",
		"config", "configuration", "settings", "preferences", "options",
		"service", "services", "server", "client", "proxy", "gateway",
		"database", "db", "cache", "storage", "backup", "archive",
		"log", "logs", "debug", "error", "warning", "info", "trace",
		"monitor", "monitoring", "metrics", "stats", "analytics",
		"deploy", "deployment", "release", "version", "build", "ci", "cd",
	}
	
	for _, word := range systemWords {
		c.systemWords.Set(word, true)
	}
}

func (c *DefaultNameClassifier) initializeCommonWords() {
	// Common English words that might be valuable
	commonWords := []string{
		"the", "and", "for", "are", "but", "not", "you", "all", "can", "had",
		"her", "was", "one", "our", "out", "day", "get", "has", "him", "his",
		"how", "man", "new", "now", "old", "see", "two", "way", "who", "boy",
		"did", "its", "let", "put", "say", "she", "too", "use", "dad", "mom",
		"big", "end", "far", "fun", "got", "lot", "may", "run", "sun", "top",
		"car", "dog", "cat", "home", "love", "life", "time", "work", "world",
		"money", "power", "peace", "happy", "music", "video", "photo", "news",
		"game", "games", "play", "sport", "food", "drink", "coffee", "tea",
		"book", "books", "read", "write", "learn", "teach", "school", "study",
		"art", "design", "color", "blue", "red", "green", "black", "white",
		"gold", "silver", "diamond", "star", "moon", "sun", "earth", "fire",
		"water", "air", "tree", "flower", "garden", "house", "city", "country",
	}
	
	for _, word := range commonWords {
		c.commonWords.Set(word, true)
	}
}

// Configuration methods

func (c *DefaultNameClassifier) SetShortNameThreshold(threshold int) {
	c.shortNameThreshold = threshold
}

func (c *DefaultNameClassifier) SetMinimumBid(amount int64) {
	c.minimumBid = amount
}

func (c *DefaultNameClassifier) AddSystemWord(word string) {
	c.systemWords.Set(strings.ToLower(word), true)
}

func (c *DefaultNameClassifier) AddTrademarkWord(word string) {
	c.trademarkWords.Set(strings.ToLower(word), true)
}

func (c *DefaultNameClassifier) AddCommonWord(word string) {
	c.commonWords.Set(strings.ToLower(word), true)
}
