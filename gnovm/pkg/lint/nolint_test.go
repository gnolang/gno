package lint

import "testing"

func TestNolintParser_Parse(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantLines []int
		wantRules map[int][]string
	}{
		{
			name:      "no nolint comments",
			source:    "package main\n\nfunc main() {}\n",
			wantLines: nil,
			wantRules: map[int][]string{},
		},
		{
			name:      "simple nolint",
			source:    "package main\n\n//nolint\nvar x int\n",
			wantLines: []int{3},
			wantRules: map[int][]string{3: nil},
		},
		{
			name:      "nolint with rule",
			source:    "package main\n\n//nolint:AVL001\nvar x int\n",
			wantLines: []int{3},
			wantRules: map[int][]string{3: {"AVL001"}},
		},
		{
			name:      "nolint with multiple rules",
			source:    "package main\n\n//nolint:AVL001,GLOBAL001\nvar x int\n",
			wantLines: []int{3},
			wantRules: map[int][]string{3: {"AVL001", "GLOBAL001"}},
		},
		{
			name:      "nolint with space",
			source:    "package main\n\n// nolint:AVL001\nvar x int\n",
			wantLines: []int{3},
			wantRules: map[int][]string{3: {"AVL001"}},
		},
		{
			name:      "multiple nolint comments",
			source:    "package main\n\n//nolint:AVL001\nvar x int\n\n//nolint:GLOBAL001\nvar y int\n",
			wantLines: []int{3, 6},
			wantRules: map[int][]string{3: {"AVL001"}, 6: {"GLOBAL001"}},
		},
		{
			name:      "nolint with leading whitespace",
			source:    "package main\n\n\t//nolint:AVL001\nvar x int\n",
			wantLines: []int{3},
			wantRules: map[int][]string{3: {"AVL001"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewNolintParser(tt.source)

			for _, line := range tt.wantLines {
				if _, ok := p.byLine[line]; !ok {
					t.Errorf("expected directive at line %d", line)
				}
			}

			for line, wantRules := range tt.wantRules {
				d, ok := p.byLine[line]
				if !ok {
					continue
				}
				if len(d.Rules) != len(wantRules) {
					t.Errorf("line %d: got %d rules, want %d", line, len(d.Rules), len(wantRules))
					continue
				}
				for i, rule := range wantRules {
					if d.Rules[i] != rule {
						t.Errorf("line %d: rule[%d] = %v, want %v", line, i, d.Rules[i], rule)
					}
				}
			}
		})
	}
}

func TestNolintParser_IsSuppressed(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		line     int
		ruleID   string
		expected bool
	}{
		{
			name:     "no nolint - not suppressed",
			source:   "package main\n\nvar x int\n",
			line:     3,
			ruleID:   "AVL001",
			expected: false,
		},
		{
			name:     "nolint all - suppressed",
			source:   "package main\n\n//nolint\nvar x int\n",
			line:     4, // issue is on line 4, nolint on line 3
			ruleID:   "AVL001",
			expected: true,
		},
		{
			name:     "nolint specific rule - matching - suppressed",
			source:   "package main\n\n//nolint:AVL001\nvar x int\n",
			line:     4,
			ruleID:   "AVL001",
			expected: true,
		},
		{
			name:     "nolint specific rule - not matching - not suppressed",
			source:   "package main\n\n//nolint:GLOBAL001\nvar x int\n",
			line:     4,
			ruleID:   "AVL001",
			expected: false,
		},
		{
			name:     "nolint multiple rules - matching - suppressed",
			source:   "package main\n\n//nolint:AVL001,GLOBAL001\nvar x int\n",
			line:     4,
			ruleID:   "GLOBAL001",
			expected: true,
		},
		{
			name:     "nolint multiple rules - not matching - not suppressed",
			source:   "package main\n\n//nolint:AVL001,GLOBAL001\nvar x int\n",
			line:     4,
			ruleID:   "OTHER001",
			expected: false,
		},
		{
			name:     "nolint on wrong line - not suppressed",
			source:   "package main\n\n//nolint:AVL001\nvar x int\nvar y int\n",
			line:     5, // issue on line 5, nolint on line 3 (covers line 4 only)
			ruleID:   "AVL001",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewNolintParser(tt.source)
			got := p.IsSuppressed(tt.line, tt.ruleID)
			if got != tt.expected {
				t.Errorf("IsSuppressed(%d, %q) = %v, want %v", tt.line, tt.ruleID, got, tt.expected)
			}
		})
	}
}

func TestNolintParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty source", ""},
		{"only newlines", "\n\n\n"},
		{"regular comment", "// this is a comment\n"},
		{"partial nolint", "//nolint"},
		{"nolint in middle of line", "var x int //nolint"}, // should not match (not at start)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			p := NewNolintParser(tt.source)
			_ = p.IsSuppressed(1, "AVL001")
		})
	}
}
