// Package gitsource derives a gnomod.toml [source] section from the git
// checkout a module lives in. It shells out to git and is meant for
// developer tooling only, never for anything that runs on chain.
package gitsource

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
)

// Detect returns the repository and path of the module in dir, from the
// "origin" remote of the enclosing git checkout. Revision is left empty: a
// committed one is stale one commit later, so it belongs to the tool that
// publishes the module.
func Detect(dir string) (gnomod.Source, error) {
	remote, err := git(dir, "remote", "get-url", "origin")
	if err != nil {
		return gnomod.Source{}, err
	}
	repo, err := NormalizeRemote(remote)
	if err != nil {
		return gnomod.Source{}, err
	}
	top, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return gnomod.Source{}, err
	}
	rel, err := relPath(top, dir)
	if err != nil {
		return gnomod.Source{}, err
	}
	src := gnomod.Source{Repository: repo, Path: rel}
	if err := src.Validate(); err != nil {
		return gnomod.Source{}, err
	}
	return src, nil
}

var scpLike = regexp.MustCompile(`^(?:[\w.-]+@)?([\w.-]+):([^/].*)$`)

// NormalizeRemote turns a git remote URL into the https URL of the
// repository: `git@github.com:o/r.git`, `ssh://git@github.com/o/r` and
// `https://token@github.com/o/r.git` all become `https://github.com/o/r`.
// Credentials are always dropped, since the result ends up on chain.
func NormalizeRemote(remote string) (string, error) {
	remote = strings.TrimSpace(remote)
	var host, p string
	if m := scpLike.FindStringSubmatch(remote); m != nil && !strings.Contains(remote, "://") {
		host, p = m[1], m[2]
	} else {
		u, err := url.Parse(remote)
		if err != nil {
			return "", fmt.Errorf("parse remote %q: %w", remote, err)
		}
		switch u.Scheme {
		case "https", "http", "ssh", "git":
		default:
			return "", fmt.Errorf("remote %q: unsupported scheme %q", remote, u.Scheme)
		}
		host, p = u.Hostname(), u.Path
	}
	p = strings.Trim(strings.TrimSuffix(strings.Trim(p, "/"), ".git"), "/")
	if host == "" || p == "" {
		return "", fmt.Errorf("remote %q: cannot derive a repository URL", remote)
	}
	return "https://" + host + "/" + p, nil
}

func relPath(top, dir string) (string, error) {
	// git reports the toplevel with symlinks resolved; resolve dir the same
	// way so a module under a symlinked checkout does not come out as "../..".
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if d, err := filepath.EvalSymlinks(absDir); err == nil {
		absDir = d
	}
	if t, err := filepath.EvalSymlinks(top); err == nil {
		top = t
	}
	rel, err := filepath.Rel(top, absDir)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if rel == "." {
		rel = ""
	}
	return rel, nil
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(out)), nil
}
