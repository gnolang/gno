package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		{"_test", "test"},
		{"_1ab", "d1ab"},
		{"__abc", "abc"},
		{"--leading-dash", "leading_dash"},
		{"123", "app"},
		{"---", "app"},
		{"", "app"},
		{"_", "app"},
		{"MIXED-Case_123", "mixed_case_123"},
		{"my.proj", "my_proj"},
		{"weird name with spaces", "weird_name_with_spaces"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitizePathSegment(tc.in))
		})
	}
}

// TestGeneratedPath_ProducesValidPath asserts that for every input basename,
// the generated pkgPath satisfies gnolang.IsUserlib. Inputs cover hyphens,
// mixed case, digit-leading, leading underscore, and non-alphanumeric chars
// — all of which gno's Re_name rejects unsanitized.
func TestGeneratedPath_ProducesValidPath(t *testing.T) {
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
			path := generatedPath(cfg, filepath.Join(t.TempDir(), base))
			assert.True(t, gnolang.IsUserlib(path),
				"generated path %q must be a valid userlib path", path)
		})
	}
}

// TestDetectLocalPackage pins the classification contract for package-dir
// candidates: a parseable gnomod.toml wins; a genuinely absent one falls
// back to the generated /r/dev/ path when the dir reads as a package; an
// INVALID gnomod.toml is an error, not "missing" — deploying such a dir
// under a generated name would hide the user's mistake.
func TestDetectLocalPackage(t *testing.T) {
	cfg := &AppConfig{chainDomain: "gno.land"}

	write := func(t *testing.T, files map[string]string) string {
		t.Helper()
		dir := filepath.Join(t.TempDir(), "myrealm")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		for name, body := range files {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
		}
		return dir
	}

	t.Run("gnomod module wins", func(t *testing.T) {
		dir := write(t, map[string]string{
			"gnomod.toml": "module = \"gno.land/r/me/realm\"\n",
			"realm.gno":   "package realm\n",
		})
		path, hasGnoMod, err := detectLocalPackage(cfg, dir)
		require.NoError(t, err)
		assert.True(t, hasGnoMod)
		assert.Equal(t, "gno.land/r/me/realm", path)
	})

	t.Run("missing gnomod generates path", func(t *testing.T) {
		dir := write(t, map[string]string{"realm.gno": "package myrealm\n"})
		path, hasGnoMod, err := detectLocalPackage(cfg, dir)
		require.NoError(t, err)
		assert.False(t, hasGnoMod)
		assert.Equal(t, "gno.land/r/dev/myrealm", path)
	})

	t.Run("no gno files errors", func(t *testing.T) {
		dir := write(t, map[string]string{"README.md": "hi\n"})
		_, _, err := detectLocalPackage(cfg, dir)
		assert.Error(t, err)
	})

	t.Run("invalid gnomod errors", func(t *testing.T) {
		dir := write(t, map[string]string{
			"gnomod.toml": "this is not toml [",
			"realm.gno":   "package myrealm\n",
		})
		_, _, err := detectLocalPackage(cfg, dir)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "gnomod.toml",
			"error must point at the broken gnomod.toml, not claim it is missing")
	})
}
