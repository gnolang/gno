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

// sanitizePathSegment lower-cases s, replaces every char outside [a-z0-9_]
// with `_`, and ensures the result matches gno's Re_name regex (optional
// leading `_`, then [a-z], then [a-z0-9_]*). Falls back to "app" when the
// input has no letters; prepends `d` when needed to satisfy the
// "must start with a letter" rule.
func sanitizePathSegment(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	for i := 0; i < len(out); i++ {
		if out[i] >= 'a' && out[i] <= 'z' {
			if i == 0 || (i == 1 && out[0] == '_') {
				return out
			}
			return "d" + out
		}
	}
	return "app"
}
