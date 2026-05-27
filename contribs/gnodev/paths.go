package main

import (
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

// guessPath returns the import path for dir. It first tries to read
// gnomod.toml; if absent, it derives a path under the chain domain's
// /r/dev/ namespace from the directory base name, sanitized to match
// gno's Re_name regex (lowercase letters/digits/underscore only,
// must start with a letter or `_<letter>`).
func guessPath(cfg *AppConfig, dir string) (path string) {
	if path, ok := guessPathGnoMod(dir); ok {
		return path
	}
	return gopath.Join(cfg.chainDomain, "/r/dev/", sanitizePathSegment(filepath.Base(dir)))
}

func guessPathGnoMod(dir string) (path string, ok bool) {
	modfile, err := gnomod.ParseDir(dir)
	if err != nil {
		return "", false
	}
	return modfile.Module, true
}

// sanitizePathSegment lower-cases s, replaces every non-alphanumeric rune
// with `_`, collapses runs of `_`, and trims leading `_`. The output matches
// gno's Re_name regex: must start with `[a-z]`, and `_` separators may only
// appear between alphanumerics. Falls back to "app" when no letters or
// digits remain; prepends `d` when the result starts with a digit.
func sanitizePathSegment(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	prevSep := true // suppresses leading separators
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevSep = false
		default:
			if !prevSep {
				b.WriteByte('_')
				prevSep = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "_")
	if !strings.ContainsFunc(out, func(r rune) bool { return r >= 'a' && r <= 'z' }) {
		return "app"
	}
	if out[0] >= '0' && out[0] <= '9' {
		return "d" + out
	}
	return out
}
