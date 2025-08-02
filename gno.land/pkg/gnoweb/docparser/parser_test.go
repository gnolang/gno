package docparser

import (
	"strings"
	"testing"
)

func TestParseDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []DocBlock
		wantErr  bool
	}{
		{
			name:     "empty documentation",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name: "simple text",
			input: "This is a simple text block.",
			expected: []DocBlock{
				{Type: "text", Content: "This is a simple text block."},
			},
			wantErr: false,
		},
		{
			name: "text with code block",
			input: `This is some text.

    func example() {
        fmt.Println("Hello, World!")
    }

More text here.`,
			expected: []DocBlock{
				{Type: "text", Content: "This is some text."},
				{Type: "code", Content: `func example() {
    fmt.Println("Hello, World!")
}`},
				{Type: "text", Content: "More text here."},
			},
			wantErr: false,
		},
		{
			name: "code block with tabs",
			input: `Text before.

	func tabbed() {
		return true
	}

Text after.`,
			expected: []DocBlock{
				{Type: "text", Content: "Text before."},
				{Type: "code", Content: `func tabbed() {
    return true
}`},
				{Type: "text", Content: "Text after."},
			},
			wantErr: false,
		},
		{
			name: "multiple code blocks",
			input: `First text.

    func first() {
        return 1
    }

Second text.

    func second() {
        return 2
    }

Third text.`,
			expected: []DocBlock{
				{Type: "text", Content: "First text."},
				{Type: "code", Content: `func first() {
    return 1
}`},
				{Type: "text", Content: "Second text."},
				{Type: "code", Content: `func second() {
    return 2
}`},
				{Type: "text", Content: "Third text."},
			},
			wantErr: false,
		},
		{
			name: "code block with empty lines",
			input: `Text before.

    func example() {
        // comment
        
        return true
    }

Text after.`,
			expected: []DocBlock{
				{Type: "text", Content: "Text before."},
				{Type: "code", Content: `func example() {
    // comment

    return true
}`},
				{Type: "text", Content: "Text after."},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDocumentation(tt.input)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d blocks, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("missing block %d", i)
					continue
				}
				
				actual := result[i]
				if actual.Type != expected.Type {
					t.Errorf("block %d: expected type %q, got %q", i, expected.Type, actual.Type)
				}
				
				if actual.Content != expected.Content {
					t.Errorf("block %d: expected content %q, got %q", i, expected.Content, actual.Content)
				}
			}
		})
	}
}

func TestParseDocumentationWithConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   ParserConfig
		expected []DocBlock
		wantErr  bool
	}{
		{
			name: "custom tab width",
			input: `Text.

	func example() {
		return true
	}`,
			config: ParserConfig{TabWidth: 2, MaxDocSize: 100000},
			expected: []DocBlock{
				{Type: "text", Content: "Text."},
				{Type: "code", Content: `func example() {
  return true
}`},
			},
			wantErr: false,
		},
		{
			name:     "documentation too large",
			input:    strings.Repeat("a", 100001),
			config:   ParserConfig{MaxDocSize: 100000},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDocumentationWithConfig(tt.input, tt.config)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d blocks, got %d", len(tt.expected), len(result))
				return
			}
			
			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("missing block %d", i)
					continue
				}
				
				actual := result[i]
				if actual.Type != expected.Type {
					t.Errorf("block %d: expected type %q, got %q", i, expected.Type, actual.Type)
				}
				
				if actual.Content != expected.Content {
					t.Errorf("block %d: expected content %q, got %q", i, expected.Content, actual.Content)
				}
			}
		})
	}
}

func TestIsIndented(t *testing.T) {
	tests := []struct {
		input    string
		tabWidth int
		expected bool
	}{
		{"", 4, false},
		{"text", 4, false},
		{"    text", 4, true},
		{"\ttext", 4, true},
		{"   text", 4, false}, // 3 spaces, not 4
		{"     text", 4, true}, // 5 spaces
		{"    ", 4, true}, // just spaces
		{"\t", 4, true}, // just tab
		{"  text", 2, true}, // 2 spaces with tabWidth=2
		{" text", 2, false}, // 1 space with tabWidth=2
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isIndented(tt.input, tt.tabWidth)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCountIndentation(t *testing.T) {
	tests := []struct {
		input    string
		tabWidth int
		expected int
	}{
		{"", 4, 0},
		{"text", 4, 0},
		{"    text", 4, 4},
		{"\ttext", 4, 4},
		{"   text", 4, 3},
		{"     text", 4, 5},
		{"    ", 4, 4},
		{"\t", 4, 4},
		{"  \t  text", 4, 8}, // 2 spaces + tab (4) + 2 spaces = 8
		{"  text", 2, 2}, // 2 spaces with tabWidth=2
		{"\ttext", 2, 2}, // tab with tabWidth=2
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := countIndentation(tt.input, tt.tabWidth)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Benchmark tests for performance
func BenchmarkParseDocumentation(b *testing.B) {
	doc := `This is a documentation with multiple code blocks.

    func example1() {
        fmt.Println("Hello")
    }

More text here.

    func example2() {
        return true
    }

Final text.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDocumentation(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDocumentationLarge(b *testing.B) {
	// Create a large documentation with many code blocks
	var lines []string
	lines = append(lines, "Large documentation with many code blocks.")
	
	for i := 0; i < 100; i++ {
		lines = append(lines, "")
		lines = append(lines, "    func example"+string(rune(i+'0'))+"() {")
		lines = append(lines, "        return "+string(rune(i+'0')))
		lines = append(lines, "    }")
		lines = append(lines, "")
		lines = append(lines, "Text between code blocks.")
	}
	
	doc := strings.Join(lines, "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDocumentation(doc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDocumentationReader(b *testing.B) {
	doc := `This is a documentation with multiple code blocks.

    func example1() {
        fmt.Println("Hello")
    }

More text here.

    func example2() {
        return true
    }

Final text.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseDocumentationReader(strings.NewReader(doc))
		if err != nil {
			b.Fatal(err)
		}
	}
} 