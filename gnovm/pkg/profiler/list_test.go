package profiler

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteFunctionList_PartialMatch(t *testing.T) {
	tests := []struct {
		name          string
		funcName      string
		samples       []ProfileSample
		expectFound   bool
		expectMatches []string // Expected function names in output
	}{
		{
			name:     "exact match",
			funcName: "gno.land/p/demo/ufmt.Sprintf",
			samples: []ProfileSample{
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/ufmt.Sprintf",
						File:     "ufmt.gno",
						Line:     10,
					}},
					Value: []int64{1, 1000},
				},
			},
			expectFound:   true,
			expectMatches: []string{"gno.land/p/demo/ufmt.Sprintf"},
		},
		{
			name:     "partial match - function name only",
			funcName: "Sprintf",
			samples: []ProfileSample{
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/ufmt.Sprintf",
						File:     "ufmt.gno",
						Line:     10,
					}},
					Value: []int64{1, 1000},
				},
			},
			expectFound:   true,
			expectMatches: []string{"gno.land/p/demo/ufmt.Sprintf"},
		},
		{
			name:     "multiple matches",
			funcName: "Sprintf",
			samples: []ProfileSample{
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/ufmt.Sprintf",
						File:     "ufmt.gno",
						Line:     10,
					}},
					Value: []int64{1, 1000},
				},
				{
					Location: []ProfileLocation{{
						Function: "fmt.Sprintf",
						File:     "fmt.gno",
						Line:     20,
					}},
					Value: []int64{1, 2000},
				},
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/other.Sprintf",
						File:     "other.gno",
						Line:     30,
					}},
					Value: []int64{1, 3000},
				},
			},
			expectFound:   true,
			expectMatches: []string{"gno.land/p/demo/ufmt.Sprintf", "fmt.Sprintf", "gno.land/p/demo/other.Sprintf"},
		},
		{
			name:     "no match",
			funcName: "NonExistent",
			samples: []ProfileSample{
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/ufmt.Sprintf",
						File:     "ufmt.gno",
						Line:     10,
					}},
					Value: []int64{1, 1000},
				},
			},
			expectFound:   false,
			expectMatches: []string{},
		},
		{
			name:     "case sensitive match",
			funcName: "sprintf", // lowercase
			samples: []ProfileSample{
				{
					Location: []ProfileLocation{{
						Function: "gno.land/p/demo/ufmt.Sprintf",
						File:     "ufmt.gno",
						Line:     10,
					}},
					Value: []int64{1, 1000},
				},
			},
			expectFound:   false,
			expectMatches: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Profile{
				Samples: tt.samples,
			}

			var buf bytes.Buffer
			err := p.WriteFunctionList(&buf, tt.funcName, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Check if function was found
			if tt.expectFound {
				if strings.Contains(output, "No samples found") {
					t.Errorf("expected to find function %q, but got no matches", tt.funcName)
				}

				// Check each expected match
				for _, expectedFunc := range tt.expectMatches {
					if !strings.Contains(output, expectedFunc) {
						t.Errorf("expected output to contain %q, but it didn't", expectedFunc)
					}
				}

				// Check for "ROUTINE" headers for multiple matches
				if len(tt.expectMatches) > 1 {
					routineCount := strings.Count(output, "ROUTINE ========================")
					if routineCount != len(tt.expectMatches) {
						t.Errorf("expected %d ROUTINE sections, got %d", len(tt.expectMatches), routineCount)
					}
				}
			} else {
				if !strings.Contains(output, "No samples found") {
					t.Errorf("expected 'No samples found' message, but found matches")
				}
			}
		})
	}
}

func TestWriteFunctionList_MultipleMatches_Formatting(t *testing.T) {
	p := &Profile{
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{{
					Function: "gno.land/p/demo/ufmt.Sprintf",
					File:     "ufmt.gno",
					Line:     10,
				}},
				Value: []int64{1, 1000},
			},
			{
				Location: []ProfileLocation{{
					Function: "fmt.Sprintf",
					File:     "fmt.gno",
					Line:     20,
				}},
				Value: []int64{1, 2000},
			},
		},
	}

	var buf bytes.Buffer
	err := p.WriteFunctionList(&buf, "Sprintf", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that both functions are displayed with proper formatting
	expectedSections := []string{
		"ROUTINE ======================== gno.land/p/demo/ufmt.Sprintf in ufmt.gno",
		"ROUTINE ======================== fmt.Sprintf in fmt.gno",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("expected output to contain section:\n%s", section)
		}
	}

	// Check that sections are separated properly
	if !strings.Contains(output, "\n\n") {
		t.Error("expected sections to be separated by blank lines")
	}
}

func TestWriteFunctionList_WithSourceCode(t *testing.T) {
	// Mock store with source code
	store := &mockStore{
		files: map[string]string{
			"demo/ufmt/ufmt.gno": `package ufmt

func Sprintf(format string, args ...interface{}) string {
	// implementation
	return ""
}`,
		},
	}

	p := &Profile{
		Samples: []ProfileSample{
			{
				Location: []ProfileLocation{{
					Function: "gno.land/p/demo/ufmt.Sprintf",
					File:     "demo/ufmt/ufmt.gno",
					Line:     3,
				}},
				Value: []int64{1, 1000},
			},
		},
	}

	var buf bytes.Buffer
	err := p.WriteFunctionList(&buf, "Sprintf", store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Check that source code is displayed
	if !strings.Contains(output, "func Sprintf(format string") {
		t.Error("expected source code to be displayed")
	}

	// Check line numbers are shown
	if !strings.Contains(output, "3:") {
		t.Error("expected line numbers to be shown")
	}
}

// mockStore implements Store interface for testing
type mockStore struct {
	files map[string]string
}

func (m *mockStore) GetMemFile(pkgPath, name string) *MemFile {
	// Try various combinations to find the file
	fullPath := pkgPath + "/" + name
	if content, ok := m.files[fullPath]; ok {
		return &MemFile{
			Name: name,
			Body: content,
		}
	}

	// Try with just the path as provided
	if content, ok := m.files[pkgPath]; ok {
		return &MemFile{
			Name: name,
			Body: content,
		}
	}

	return nil
}
