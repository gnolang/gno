package main

import (
	gopath "path"
	"path/filepath"
	"regexp"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

var reInvalidChar = regexp.MustCompile(`[^\w_-]`)

// guessPath returns the import path for dir. It first tries to read
// gnomod.toml; if absent, it derives a path under the chain domain's
// /r/dev/ namespace from the directory base name.
func guessPath(cfg *AppConfig, dir string) (path string) {
	if path, ok := guessPathGnoMod(dir); ok {
		return path
	}

	rname := reInvalidChar.ReplaceAllString(filepath.Base(dir), "-")
	return gopath.Join(cfg.chainDomain, "/r/dev/", rname)
}

func guessPathGnoMod(dir string) (path string, ok bool) {
	modfile, err := gnomod.ParseDir(dir)
	if err != nil {
		return "", false
	}
	return modfile.Module, true
}
