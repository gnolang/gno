package main

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
)

func TestSanitizePathSegment(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"myproj", "myproj"},
		{"gnodev-smoke", "gnodev_smoke"},
		{"My-App", "my_app"},
		{"1stproj", "d1stproj"},
		{"_test", "_test"},
		{"_1ab", "d_1ab"},
		{"__abc", "d__abc"},
		{"123", "app"},
		{"---", "app"},
		{"", "app"},
		{"_", "app"},
		{"MIXED-Case_123", "mixed_case_123"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitizePathSegment(tc.in))
		})
	}
}

// TestGuessPath_NoGnoModProducesValidPath ensures every dir basename that
// the regex sanitizer accepts produces a pkgPath that gno's IsUserlib
// validator accepts. Regression for the bug where -<hyphen> in dir name
// caused gnodev startup to panic.
func TestGuessPath_NoGnoModProducesValidPath(t *testing.T) {
	cfg := &AppConfig{chainDomain: "gno.land"}
	for _, base := range []string{
		"myproj",
		"gnodev-smoke",
		"My-App",
		"1stproj",
		"_test",
		"123",
		"weird name with spaces",
		"my.proj",
		"--leading-dash",
	} {
		t.Run(base, func(t *testing.T) {
			tempDir := filepath.Join(t.TempDir(), base)
			// We don't actually create the dir; guessPath only reads
			// gnomod.toml or falls back to base name handling.
			path := guessPath(cfg, tempDir)
			assert.True(t, gnolang.IsUserlib(path),
				"guessed path %q must be a valid userlib path", path)
		})
	}
}
