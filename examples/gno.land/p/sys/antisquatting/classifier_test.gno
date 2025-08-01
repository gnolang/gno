package antisquatting

import (
	"testing"
	"time"
)

func TestDefaultNameClassifier_IsHighValue(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	tests := []struct {
		name     string
		expected bool
		reason   string
	}{
		// Short names
		{"a", true, "single character"},
		{"ab", true, "two characters"},
		{"abc", true, "three characters"},
		{"abcd", false, "four characters (threshold)"},
		{"abcde", false, "five characters"},
		
		// System words
		{"admin", true, "system reserved word"},
		{"root", true, "system reserved word"},
		{"system", true, "system reserved word"},
		{"gov", true, "system reserved word"},
		{"dao", true, "system reserved word"},
		{"api", true, "system reserved word"},
		{"www", true, "system reserved word"},
		
		// Common words
		{"the", true, "common word"},
		{"and", true, "common word"},
		{"home", true, "common word"},
		{"love", true, "common word"},
		{"money", true, "common word"},
		{"power", true, "common word"},
		
		// High-value patterns
		{"aaa", true, "repeated characters"},
		{"111", true, "repeated numbers"},
		{"abc", true, "sequential characters"},
		{"123", true, "sequential numbers"},
		{"aba", true, "palindrome"},
		{"121", true, "palindrome"},
		
		// Regular names
		{"john", false, "regular name"},
		{"alice", false, "regular name"},
		{"user123", false, "regular name"},
		{"myname", false, "regular name"},
		{"randomuser", false, "regular name"},
		
		// Invalid names
		{"", false, "empty name"},
		{"user@domain", false, "invalid characters"},
		{"user-name", false, "invalid characters"},
	}
	
	for _, test := range tests {
		result := classifier.IsHighValue(test.name)
		if result != test.expected {
			t.Errorf("IsHighValue(%q) = %v, expected %v (%s)", 
				test.name, result, test.expected, test.reason)
		}
	}
}

func TestDefaultNameClassifier_GetMinimumBid(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	tests := []struct {
		name        string
		expectedMin int64
		reason      string
	}{
		// Single character - highest multiplier
		{"a", classifier.minimumBid * classifier.shortNameMultiplier * 100, "single character"},
		
		// Two characters
		{"ab", classifier.minimumBid * classifier.shortNameMultiplier * 50, "two characters"},
		
		// Three characters
		{"abc", classifier.minimumBid * classifier.shortNameMultiplier * 20, "three characters"},
		
		// System words
		{"admin", classifier.minimumBid * classifier.systemNameMultiplier, "system word"},
		{"root", classifier.minimumBid * classifier.systemNameMultiplier, "system word"},
		
		// Regular high-value names
		{"home", classifier.minimumBid, "common word"},
		{"love", classifier.minimumBid, "common word"},
		
		// Regular names (should still return minimum)
		{"regularname", classifier.minimumBid, "regular name"},
	}
	
	for _, test := range tests {
		result := classifier.GetMinimumBid(test.name)
		if result != test.expectedMin {
			t.Errorf("GetMinimumBid(%q) = %d, expected %d (%s)", 
				test.name, result, test.expectedMin, test.reason)
		}
	}
}

func TestDefaultNameClassifier_GetAuctionDuration(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	tests := []struct {
		name     string
		expected time.Duration
		reason   string
	}{
		// System words get longer duration
		{"admin", classifier.systemNameCommitDuration, "system word"},
		{"root", classifier.systemNameCommitDuration, "system word"},
		
		// Short names get longer duration
		{"a", classifier.shortNameCommitDuration, "short name"},
		{"ab", classifier.shortNameCommitDuration, "short name"},
		{"abc", classifier.shortNameCommitDuration, "short name"},
		
		// Regular names get default duration
		{"home", classifier.defaultCommitDuration, "regular high-value name"},
		{"regularname", classifier.defaultCommitDuration, "regular name"},
	}
	
	for _, test := range tests {
		result := classifier.GetAuctionDuration(test.name)
		if result != test.expected {
			t.Errorf("GetAuctionDuration(%q) = %v, expected %v (%s)", 
				test.name, result, test.expected, test.reason)
		}
	}
}

func TestDefaultNameClassifier_GetMinimumDeposit(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	tests := []struct {
		name     string
		expected int64
		reason   string
	}{
		// System words require higher deposits
		{"admin", classifier.systemNameDeposit, "system word"},
		{"root", classifier.systemNameDeposit, "system word"},
		
		// Short names require higher deposits
		{"a", classifier.shortNameDeposit, "short name"},
		{"ab", classifier.shortNameDeposit, "short name"},
		{"abc", classifier.shortNameDeposit, "short name"},
		
		// Regular names get default deposit
		{"home", classifier.defaultDeposit, "regular high-value name"},
		{"regularname", classifier.defaultDeposit, "regular name"},
	}
	
	for _, test := range tests {
		result := classifier.GetMinimumDeposit(test.name)
		if result != test.expected {
			t.Errorf("GetMinimumDeposit(%q) = %d, expected %d (%s)", 
				test.name, result, test.expected, test.reason)
		}
	}
}

func TestDefaultNameClassifier_PatternDetection(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	// Test repeated character detection
	repeatedTests := []struct {
		name     string
		expected bool
	}{
		{"aaa", true},
		{"111", true},
		{"aaaa", true},
		{"abc", false},
		{"121", false},
		{"aa", false}, // Too short
	}
	
	for _, test := range repeatedTests {
		result := classifier.isRepeatedChar(test.name)
		if result != test.expected {
			t.Errorf("isRepeatedChar(%q) = %v, expected %v", 
				test.name, result, test.expected)
		}
	}
	
	// Test sequential detection
	sequentialTests := []struct {
		name     string
		expected bool
	}{
		{"abc", true},
		{"123", true},
		{"cba", true}, // Descending
		{"321", true}, // Descending
		{"abcd", true},
		{"ace", false}, // Not sequential
		{"ab", false},  // Too short
	}
	
	for _, test := range sequentialTests {
		result := classifier.isSequential(test.name)
		if result != test.expected {
			t.Errorf("isSequential(%q) = %v, expected %v", 
				test.name, result, test.expected)
		}
	}
	
	// Test palindrome detection
	palindromeTests := []struct {
		name     string
		expected bool
	}{
		{"aba", true},
		{"121", true},
		{"abcba", true},
		{"12321", true},
		{"abc", false},
		{"123", false},
		{"ab", false}, // Too short for palindrome check
	}
	
	for _, test := range palindromeTests {
		result := classifier.isPalindrome(test.name)
		if result != test.expected {
			t.Errorf("isPalindrome(%q) = %v, expected %v", 
				test.name, result, test.expected)
		}
	}
}

func TestDefaultNameClassifier_Configuration(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	// Test configuration changes
	originalThreshold := classifier.shortNameThreshold
	classifier.SetShortNameThreshold(5)
	
	if classifier.shortNameThreshold != 5 {
		t.Errorf("SetShortNameThreshold failed: got %d, expected 5", 
			classifier.shortNameThreshold)
	}
	
	// Test that the change affects classification
	if !classifier.IsHighValue("abcd") {
		t.Error("Name 'abcd' should be high-value with threshold 5")
	}
	
	// Reset
	classifier.SetShortNameThreshold(originalThreshold)
	
	// Test minimum bid configuration
	originalBid := classifier.minimumBid
	classifier.SetMinimumBid(2_000_000)
	
	if classifier.minimumBid != 2_000_000 {
		t.Errorf("SetMinimumBid failed: got %d, expected 2000000", 
			classifier.minimumBid)
	}
	
	// Reset
	classifier.SetMinimumBid(originalBid)
	
	// Test adding custom words
	classifier.AddSystemWord("custom")
	if !classifier.IsHighValue("custom") {
		t.Error("Custom system word should be high-value")
	}
	
	classifier.AddTrademarkWord("brand")
	if !classifier.IsHighValue("brand") {
		t.Error("Custom trademark word should be high-value")
	}
	
	classifier.AddCommonWord("special")
	if !classifier.IsHighValue("special") {
		t.Error("Custom common word should be high-value")
	}
}

func TestDefaultNameClassifier_EdgeCases(t *testing.T) {
	classifier := NewDefaultNameClassifier()
	
	// Test case sensitivity
	if classifier.IsHighValue("ADMIN") != classifier.IsHighValue("admin") {
		t.Error("Classification should be case-insensitive")
	}
	
	// Test whitespace handling
	if classifier.IsHighValue(" admin ") != classifier.IsHighValue("admin") {
		t.Error("Classification should handle whitespace")
	}
	
	// Test empty and invalid inputs
	if classifier.IsHighValue("") {
		t.Error("Empty string should not be high-value")
	}
	
	// Test very long names
	longName := "verylongnamethatexceedsreasonablelimits"
	if classifier.IsHighValue(longName) {
		t.Error("Very long names should not be high-value by default")
	}
}
