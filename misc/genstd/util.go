package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// for tests
var skipExternalTools bool

func runTool(importPath string) error {
	if skipExternalTools {
		return nil
	}
	shortName := path.Base(importPath)
	gr := gitRoot()

	cmd := exec.Command(
		"go", "run", "-modfile", filepath.Join(gr, "misc/devdeps/go.mod"),
		importPath, "-w", outputFile,
	)
	_, err := cmd.Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("error executing %s: %w; output: %v", shortName, err, string(err.Stderr))
		}
		return fmt.Errorf("error executing %s: %w", shortName, err)
	}
	return nil
}

var (
	memoGitRoot string
	memoRelPath string

	dirsOnce sync.Once
)

func gitRoot() string {
	dirsOnceDo()
	return memoGitRoot
}

func relPath() string {
	dirsOnceDo()
	return memoRelPath
}

func dirsOnceDo() {
	dirsOnce.Do(func() {
		var err error
		memoGitRoot, memoRelPath, err = findDirs()
		if err != nil {
			panic(fmt.Errorf("could not determine git root: %w", err))
		}
	})
}

func findDirs() (gitRoot string, relPath string, err error) {
	wd, err := os.Getwd()
	if err != nil {
		return
	}
	// Resolve symlinks so that wd and git-returned paths can be compared.
	// os.Getwd preserves symlinks, while git resolves them (e.g. /tmp -> /private/tmp on macOS).
	if resolved, e := filepath.EvalSymlinks(wd); e == nil {
		wd = resolved
	}

	// makeRelPath computes the relative path from root to wd, with / separators.
	makeRelPath := func(root string) string {
		rp := strings.TrimPrefix(wd, root+string(filepath.Separator))
		return strings.ReplaceAll(rp, string(filepath.Separator), "/")
	}

	// Try git rev-parse --show-toplevel first; this correctly handles git worktrees.
	if out, e := exec.Command("git", "rev-parse", "--show-toplevel").Output(); e == nil {
		p := strings.TrimSpace(string(out))
		return p, makeRelPath(p), nil
	}

	// Fall back to walking up parent directories looking for a .git entry.
	p := wd
	for {
		if _, e := os.Stat(filepath.Join(p, ".git")); e == nil {
			return p, makeRelPath(p), nil
		}

		if strings.HasSuffix(p, string(filepath.Separator)) {
			return "", "", errors.New("root git not found")
		}

		p = filepath.Dir(p)
	}
}

// pkgNameFromPath derives the package name from the given path,
// unambiguously for the most part (so safe for the code generation).
//
// The path is taken and possibly shortened if it starts with a known prefix.
// For instance, github.com/gnolang/gno/stdlibs/std simply becomes "libs_std".
// "Unsafe" characters are removed (ie. invalid for go identifiers).
func pkgNameFromPath(path string) string {
	const (
		repoPrefix     = "github.com/gnolang/gno/"
		vmPrefix       = repoPrefix + "gnovm/"
		tm2Prefix      = repoPrefix + "tm2/pkg/"
		libsPrefix     = vmPrefix + "stdlibs/"
		testlibsPrefix = vmPrefix + "tests/stdlibs/"
	)

	ns := "ext"
	switch {
	case strings.HasPrefix(path, testlibsPrefix):
		ns, path = "testlibs", path[len(testlibsPrefix):]
	case strings.HasPrefix(path, libsPrefix):
		ns, path = "libs", path[len(libsPrefix):]
	case strings.HasPrefix(path, vmPrefix):
		ns, path = "vm", path[len(vmPrefix):]
	case strings.HasPrefix(path, tm2Prefix):
		ns, path = "tm2", path[len(tm2Prefix):]
	case strings.HasPrefix(path, repoPrefix):
		ns, path = "repo", path[len(repoPrefix):]
	case !strings.Contains(path, "."):
		ns = "go"
	}

	flds := strings.FieldsFunc(path, func(r rune) bool {
		return (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9')
	})
	return ns + "_" + strings.Join(flds, "_")
}

func mustUnquote(v string) string {
	s, err := strconv.Unquote(v)
	if err != nil {
		panic(fmt.Errorf("could not unquote import path literal: %s", v))
	}
	return s
}
