package main

import (
	"errors"
	"fmt"
	gopath "path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

// detectLocalPackage classifies dir as a deployable package candidate.
// A gnomod.toml module path wins; a genuinely absent gnomod.toml falls back
// to the generated /r/dev/ path, accepted only when the dir actually reads
// as a gno package. An unparseable gnomod.toml is an error — deploying such
// a dir under a generated name would hide the user's mistake.
func detectLocalPackage(cfg *AppConfig, dir string) (path string, hasGnoMod bool, err error) {
	mod, err := gnomod.ParseDir(dir)
	switch {
	case err == nil:
		return mod.Module, true, nil
	case !errors.Is(err, gnomod.ErrNoModFile):
		return "", false, fmt.Errorf("invalid gnomod.toml: %w", err)
	}

	path = generatedPath(cfg, dir)
	if _, err := gnolang.ReadMemPackage(dir, path, gnolang.MPAnyAll); err != nil {
		return "", false, fmt.Errorf("no gno package found: %w", err)
	}
	return path, false, nil
}

// generatedPath derives a module path under the chain domain's /r/dev/
// namespace from the directory base name, sanitized to match gno's Re_name
// regex (lowercase letters/digits/underscore only, must start with a letter
// or `_<letter>`).
func generatedPath(cfg *AppConfig, dir string) string {
	return gopath.Join(cfg.chainDomain, "/r/dev/", sanitizePathSegment(filepath.Base(dir)))
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
