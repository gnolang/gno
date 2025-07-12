package coverage

import "testing"

func TestVisualize_GlobToRegex(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{"*.gno", "^.*\\.gno$"},
		{"test*.gno", "^test.*\\.gno$"},
		{"*test.gno", "^.*test\\.gno$"},
		{"test.gno", "^test\\.gno$"},
		{"file?.gno", "^file.\\.gno$"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := globToRegex(tt.pattern)
			t.Logf("pattern: %s, result: %s", tt.pattern, result)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestVisualize_GetLineColor(t *testing.T) {
	originalSupportsColor := supportsColor
	supportsColor = true

	tests := []struct {
		name     string
		lineNum  int
		data     *CoverageData
		expected string
	}{
		{
			name:    "executed line with count 1",
			lineNum: 10,
			data: &CoverageData{
				LineData: map[int]int{10: 1},
			},
			expected: ColorGreen,
		},
		{
			name:    "executed line with count > 1",
			lineNum: 15,
			data: &CoverageData{
				LineData: map[int]int{15: 5},
			},
			expected: ColorGreen,
		},
		{
			name:    "executable but not executed line",
			lineNum: 20,
			data: &CoverageData{
				LineData: map[int]int{20: 0},
			},
			expected: ColorRed,
		},
		{
			name:    "non-instrumented line",
			lineNum: 25,
			data: &CoverageData{
				LineData: map[int]int{10: 1, 15: 0},
			},
			expected: ColorWhite,
		},
		{
			name:    "empty coverage data",
			lineNum: 30,
			data: &CoverageData{
				LineData: map[int]int{},
			},
			expected: ColorWhite,
		},
		{
			name:     "nil coverage data",
			lineNum:  35,
			data:     nil,
			expected: ColorWhite,
		},
		{
			name:    "negative line number",
			lineNum: -1,
			data: &CoverageData{
				LineData: map[int]int{10: 1},
			},
			expected: ColorWhite,
		},
		{
			name:    "zero line number",
			lineNum: 0,
			data: &CoverageData{
				LineData: map[int]int{0: 1},
			},
			expected: ColorGreen,
		},
		{
			name:    "very large line number",
			lineNum: 999999,
			data: &CoverageData{
				LineData: map[int]int{10: 1},
			},
			expected: ColorWhite,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getLineColor(tt.lineNum, tt.data)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for line %d", tt.expected, result, tt.lineNum)
			}
		})
	}

	supportsColor = false
	data := &CoverageData{
		LineData: map[int]int{
			10: 1,
			15: 0,
		},
	}
	color := getLineColor(10, data)
	if color != "" {
		t.Errorf("Expected empty string when color not supported, got %s", color)
	}
	color = getLineColor(15, data)
	if color != "" {
		t.Errorf("Expected empty string when color not supported, got %s", color)
	}
	color = getLineColor(20, data)
	if color != "" {
		t.Errorf("Expected empty string when color not supported, got %s", color)
	}

	supportsColor = originalSupportsColor
}

func TestVisualize_GetLineIndicator(t *testing.T) {
	data := &CoverageData{
		LineData: map[int]int{
			10: 1, // executed
			15: 0, // executable but not executed
		},
	}

	// Test executed line
	indicator := getLineIndicator(10, data)
	if indicator != "✓" {
		t.Errorf("Expected ✓ for executed line, got %s", indicator)
	}

	// Test executable but not executed line
	indicator = getLineIndicator(15, data)
	if indicator != "✗" {
		t.Errorf("Expected ✗ for non-executed line, got %s", indicator)
	}

	// Test non-instrumented line
	indicator = getLineIndicator(20, data)
	if indicator != " " {
		t.Errorf("Expected space for non-instrumented line, got %s", indicator)
	}
}
