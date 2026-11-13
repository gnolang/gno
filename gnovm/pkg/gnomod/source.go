package gnomod

import (
	"fmt"
	"strconv"
	"strings"
)

// Source points at the repository holding a module's source code.
type Source struct {
	// Repository is the https URL of the repository,
	// like `https://github.com/gnolang/gno`.
	Repository string `toml:"repository,omitempty" json:"repository,omitempty"`
	// Path is the module's directory inside the repository, slash-separated
	// and relative to its root, like `examples/gno.land/p/nt/avl`.
	// Empty means the repository root.
	Path string `toml:"path,omitempty" json:"path,omitempty"`
	// Revision is the commit the module was published from. It is meant to be
	// stamped by publish tooling on the uploaded copy rather than committed,
	// since a committed value is stale as soon as the next commit lands.
	Revision string `toml:"revision,omitempty" json:"revision,omitempty"`
}

// IsZero reports whether no source is declared.
func (s Source) IsZero() bool {
	return s == Source{}
}

// maxSourceFieldLen bounds each [source] value. The fields are caller-supplied
// and stored on chain, and nothing legitimate comes close.
const maxSourceFieldLen = 256

// Validate checks the [source] section against a fixed ASCII grammar:
//
//	repository = "https://" host [ ":" port ] 1*( "/" segment )
//	host       = label 1*( "." label )         ; lowercase, at least one dot
//	label      = [a-z0-9] / [a-z0-9][a-z0-9-]*[a-z0-9]
//	port       = 1..65535, no leading zero
//	path       = segment *( "/" segment )
//	segment    = 1*[A-Za-z0-9._~-], but not "." or ".."
//	revision   = 7..64 lowercase hex digits
//
// Every value is at most maxSourceFieldLen bytes.
//
// The grammar is hand-written rather than delegated to net/url on purpose.
// A value rendered as an href must not carry a script scheme, credentials, a
// query, a fragment, a percent-encoded or literal dot segment that a browser
// would resolve out of the repository, or a non-ASCII homoglyph; and the
// result must not depend on the Go version or GODEBUG settings (url.Parse
// changed under urlstrictcolons in Go 1.26), since the keeper runs this.
//
// It is NOT part of File.Validate, so an invalid section never makes a
// gnomod.toml unreadable: read paths (gno test, gnodev, gnoweb, the keeper
// re-reading stored packages) ignore it, and Source.URL returns nothing for
// it. It runs where a section is written: AddPackage (checkGnomodConstraints
// in the vm keeper) and `gno mod init/tidy`.
//
// These rules are consensus-visible once a package is stored, so they may
// only ever be loosened. Tightening them would refuse a replayed AddPackage.
func (s Source) Validate() error {
	if s.IsZero() {
		return nil
	}
	if s.Repository == "" {
		return fmt.Errorf("'source.repository' is required when [source] is set")
	}
	// A slice, not a map, so the field reported first is stable.
	for _, f := range []struct{ name, v string }{
		{"repository", s.Repository},
		{"path", s.Path},
		{"revision", s.Revision},
	} {
		if len(f.v) > maxSourceFieldLen {
			return fmt.Errorf("'source.%s' exceeds %d bytes", f.name, maxSourceFieldLen)
		}
	}
	if _, _, err := splitRepository(s.Repository); err != nil {
		return fmt.Errorf("'source.repository' %w, got %q", err, s.Repository)
	}
	if p := s.Path; p != "" {
		if !validSegments(strings.Split(p, "/")) {
			return fmt.Errorf("'source.path' must be a clean relative slash-separated path of [A-Za-z0-9._~-] segments, got %q", p)
		}
	}
	if r := s.Revision; r != "" {
		if len(r) < 7 || len(r) > 64 || strings.Trim(r, "0123456789abcdef") != "" {
			return fmt.Errorf("'source.revision' must be a lowercase hex commit id, got %q", r)
		}
	}
	return nil
}

// splitRepository parses a repository URL per the grammar in Validate and
// returns its host (without port) and its path segments.
func splitRepository(repo string) (host string, segments []string, err error) {
	rest, ok := strings.CutPrefix(repo, "https://")
	if !ok {
		return "", nil, fmt.Errorf("must be an https URL")
	}
	hostport, p, _ := strings.Cut(rest, "/")
	host, port, hasPort := strings.Cut(hostport, ":")
	if !validHost(host) {
		return "", nil, fmt.Errorf("must have a lowercase DNS host with at least one dot")
	}
	if hasPort {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 || port[0] == '0' || strings.Trim(port, "0123456789") != "" {
			return "", nil, fmt.Errorf("has an invalid port")
		}
	}
	segments = strings.Split(p, "/")
	if p == "" || !validSegments(segments) {
		return "", nil, fmt.Errorf("must have a path of [A-Za-z0-9._~-] segments, without credentials, query, fragment, percent-encoding or dot segments")
	}
	return host, segments, nil
}

func validHost(h string) bool {
	labels := strings.Split(h, ".")
	if len(labels) < 2 {
		return false
	}
	for _, l := range labels {
		if l == "" || l[0] == '-' || l[len(l)-1] == '-' {
			return false
		}
		for i := 0; i < len(l); i++ {
			c := l[i]
			if !('a' <= c && c <= 'z' || '0' <= c && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}

func validSegments(segs []string) bool {
	for _, seg := range segs {
		if seg == "" || seg == "." || seg == ".." {
			return false
		}
		for i := 0; i < len(seg); i++ {
			c := seg[i]
			if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
				c == '.' || c == '_' || c == '~' || c == '-') {
				return false
			}
		}
	}
	return true
}

// URL returns a browsable link to the module's source, and whether that link
// is pinned to Revision. It returns "" for an invalid section.
//
// On github.com (owner/repo) and gitlab.com (group/.../repo) it points at the
// module directory at Revision, or at HEAD when none is set. On any other host
// there is no URL scheme to rely on, so it is the repository URL itself and
// never pinned.
func (s Source) URL() (link string, pinned bool) {
	if s.Repository == "" || s.Validate() != nil {
		return "", false
	}
	host, segs, _ := splitRepository(s.Repository)
	var tree string
	switch {
	case host == "github.com" && len(segs) == 2:
		tree = "/tree/"
	case host == "gitlab.com" && len(segs) >= 2:
		tree = "/-/tree/"
	default:
		return s.Repository, false
	}
	if s.Path == "" && s.Revision == "" {
		return s.Repository, false
	}
	rev := s.Revision
	if rev == "" {
		rev = "HEAD"
	}
	// A clone URL ending in .git would 404 once a tree path is appended.
	link = strings.TrimSuffix(s.Repository, ".git") + tree + rev
	if s.Path != "" {
		link += "/" + s.Path
	}
	return link, s.Revision != ""
}
