package coverage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupCoverage func() *Coverage
		location      FileLocation
		expectedHits  map[string]map[int]int
		checkLocation bool
	}{
		{
			name: "Record hit for new file and line",
			setupCoverage: func() *Coverage {
				c := New("")
				c.Enable()
				c.SetExecutableLines("testpkg/testfile.gno", map[int]bool{10: true})
				return c
			},
			location: FileLocation{
				PkgPath: "testpkg",
				File:    "testfile.gno",
				Line:    10,
				Column:  5,
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {10: 1},
			},
			checkLocation: true,
		},
		{
			name: "Increment hit count for existing line",
			setupCoverage: func() *Coverage {
				c := New("")
				c.Enable()
				c.SetExecutableLines("testpkg/testfile.gno", map[int]bool{10: true})
				// pre-record a hit
				c.RecordHit(FileLocation{
					PkgPath: "testpkg",
					File:    "testfile.gno",
					Line:    10,
				})
				return c
			},
			location: FileLocation{
				PkgPath: "testpkg",
				File:    "testfile.gno",
				Line:    10,
				Column:  5,
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {10: 2},
			},
			checkLocation: true,
		},
		{
			name: "Do not record coverage for non-executable line",
			setupCoverage: func() *Coverage {
				c := New("")
				c.Enable()
				c.SetExecutableLines("testpkg/testfile.gno", map[int]bool{10: true})
				return c
			},
			location: FileLocation{
				PkgPath: "testpkg",
				File:    "testfile.gno",
				Line:    20,
				Column:  5,
			},
			expectedHits: map[string]map[int]int{
				"testpkg/testfile.gno": {},
			},
			checkLocation: true,
		},
		{
			name: "Ignore coverage when disabled",
			setupCoverage: func() *Coverage {
				c := New("")
				c.Disable()
				return c
			},
			location: FileLocation{
				PkgPath: "testpkg",
				File:    "testfile.gno",
				Line:    10,
			},
			expectedHits:  map[string]map[int]int{},
			checkLocation: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cov := tt.setupCoverage()
			cov.RecordHit(tt.location)

			for file, expectedHits := range tt.expectedHits {
				actualHits := cov.files[file].hitLines
				assert.Equal(t, expectedHits, actualHits)
			}
		})
	}
}
