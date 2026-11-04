// Package gitsource derives a gnomod.toml [source] section from the git
// checkout a module lives in. It shells out to git and is meant for
// developer tooling only, never for anything that runs on chain.
package gitsource

import (
	"bytes"
	"fmt"
	"net/url"
	"os/exec"
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
	// --show-prefix resolves symlinks and case-insensitive filesystems the
	// way git itself does, and answers "" at the top.
	prefix, err := git(dir, "rev-parse", "--show-prefix")
	if err != nil {
		return gnomod.Source{}, err
	}
	src := gnomod.Source{Repository: repo, Path: strings.TrimSuffix(prefix, "/")}
	if err := src.Validate(); err != nil {
		return gnomod.Source{}, fmt.Errorf("%w; set [source] by hand", err)
	}
	return src, nil
}

var scpLike = regexp.MustCompile(`^(?:[\w.-]+@)?([\w.-]+):([^/].*)$`)

// NormalizeRemote turns a git remote URL into the https URL of the
// repository: `git@github.com:o/r.git`, `ssh://git@github.com/o/r` and
// `https://token@github.com/o/r.git` all become `https://github.com/o/r`.
// Credentials are always dropped, since the result ends up on chain.
//
// The port of an http(s) remote is kept, since it names the web server; the
// port of an ssh remote is not. A host without a dot is refused: it is an ssh
// alias from ~/.ssh/config, or a Windows drive letter, not a web host.
func NormalizeRemote(remote string) (string, error) {
	remote = strings.TrimSpace(remote)
	if strings.Contains(remote, `\`) {
		return "", fmt.Errorf("remote %q is a local path", remote)
	}
	var host, p string
	if m := scpLike.FindStringSubmatch(remote); m != nil && !strings.Contains(remote, "://") {
		host, p = m[1], m[2]
	} else {
		u, err := url.Parse(remote)
		if err != nil {
			return "", fmt.Errorf("parse remote %q: %w", remote, err)
		}
		switch u.Scheme {
		case "https", "http":
			host = u.Host
		case "ssh", "git":
			host = u.Hostname()
		default:
			return "", fmt.Errorf("remote %q: unsupported scheme %q", remote, u.Scheme)
		}
		// Escaped, so a %20 stays encoded and is refused by Validate rather
		// than stored as a literal space.
		p = u.EscapedPath()
	}
	host = strings.ToLower(host)
	if !strings.Contains(host, ".") {
		return "", fmt.Errorf("remote %q: host %q is not a web host (an ssh alias or a drive letter?)", remote, host)
	}
	p = strings.Trim(strings.TrimSuffix(strings.Trim(p, "/"), ".git"), "/")
	if host == "" || p == "" {
		return "", fmt.Errorf("remote %q: cannot derive a repository URL", remote)
	}
	return "https://" + host + "/" + p, nil
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
