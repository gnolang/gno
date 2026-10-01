package gnomod

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/module"
)

// Parsed gnomod.toml file.
type File struct {
	// Module is the path of the module.
	// Like `gno.land/r/path/to/module`.
	Module string `toml:"module" json:"module"`

	// Gno is the gno version string for compatibility within the gno toolchain.
	// It is intended to be set by the `gno` cli when initializing or upgrading a module.
	Gno string `toml:"gno" json:"gno"`

	// Ignore indicate that the module will be ignored by the gno toolchain but still usable in development environments.
	Ignore bool `toml:"ignore,omitempty" json:"ignore,omitempty"`

	// Draft indicates that the module isn't ready for production use.
	// Draft modules:
	// - are added to the chain at genesis time and cannot be added after.
	// - cannot be imported by other newly added modules.
	Draft bool `toml:"draft,omitempty" json:"draft,omitempty"`

	// Private indicates that the module is private.
	// Private modules:
	// - Cannot be imported by other realms.
	// - References to objects owned by this realm cannot be stored outside it.
	// - Data whose type is defined in this realm cannot be retained in other realms.
	Private bool `toml:"private,omitempty" json:"private,omitempty"`

	// Replace is a list of replace directives for the module's dependencies.
	// Each replace can link to a different online module path, or a local path.
	// If this value is set, the module cannot be added to the chain.
	Replace []Replace `toml:"replace,omitempty" json:"replace,omitempty"`

	// Source declares where the module's source code lives, so explorers can
	// link a deployed package back to it. It is informational: the chain stores
	// it as written and never fetches or verifies it.
	Source Source `toml:"source,omitempty" json:"source,omitempty"`

	// AddPkg is the addpkg section of the gnomod.toml file.
	// It is filled by the vmkeeper when a module is added.
	// It is not intended to be used offchain.
	AddPkg AddPkg `toml:"addpkg,omitempty" json:"addpkg,omitempty"`
}

type AddPkg struct {
	// Creator is the address of the creator.
	Creator string `toml:"creator,omitempty" json:"creator,omitempty"`
	// Height is the block height at which the module was added.
	Height int `toml:"height,omitempty" json:"height,omitempty"`
	// MaxDeposit is the storage-deposit ceiling that applies when the charge
	// is paid by a later message than the one that submitted the package --
	// the "inert" policy, where MsgEnablePackage pays.
	//
	// It is what the submitter declared on MsgAddPackage, or the chain default
	// as it stood at submit time if they declared nothing. Either way it is
	// fixed when the package is stored, so a later change to the chain default
	// cannot widen what the submitter is charged.
	//
	// Empty on the ordinary path: there the ceiling is used in the same
	// transaction that declared it, so nothing needs to outlive it.
	MaxDeposit string `toml:"max_deposit,omitempty" json:"max_deposit,omitempty"`
	// XXX: GnoVersion // gno version at add time?
	// XXX: Consider things like IsUsingBanker or other security-awareness flags
}

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
	// stamped by deploy tooling rather than committed, since a committed value
	// is stale as soon as the next commit lands.
	Revision string `toml:"revision,omitempty" json:"revision,omitempty"`
}

// IsZero reports whether no source is declared.
func (s Source) IsZero() bool {
	return s == Source{}
}

// maxSourceFieldLen bounds each [source] value. The fields are caller-supplied
// and stored on chain, and nothing legitimate comes close.
const maxSourceFieldLen = 256

// Validate checks the [source] section. Only https URLs without credentials,
// query or fragment are accepted, so that a value rendered as a link by an
// explorer cannot carry a script scheme or a leaked token.
func (s Source) Validate() error {
	if s.IsZero() {
		return nil
	}
	if s.Repository == "" {
		return fmt.Errorf("'source.repository' is required when [source] is set")
	}
	for name, v := range map[string]string{"repository": s.Repository, "path": s.Path, "revision": s.Revision} {
		if len(v) > maxSourceFieldLen {
			return fmt.Errorf("'source.%s' exceeds %d bytes", name, maxSourceFieldLen)
		}
		if strings.ContainsFunc(v, func(r rune) bool { return r < 0x20 || r == 0x7f }) {
			return fmt.Errorf("'source.%s' contains control characters", name)
		}
	}
	u, err := url.Parse(s.Repository)
	if err != nil {
		return fmt.Errorf("'source.repository': %w", err)
	}
	switch {
	case u.Scheme != "https":
		return fmt.Errorf("'source.repository' must be an https URL, got %q", s.Repository)
	case u.Host == "":
		return fmt.Errorf("'source.repository' has no host")
	case u.User != nil:
		return fmt.Errorf("'source.repository' must not embed credentials")
	case u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(s.Repository, "?#"):
		return fmt.Errorf("'source.repository' must not have a query or fragment")
	}
	if p := s.Path; p != "" {
		if strings.HasPrefix(p, "/") || strings.Contains(p, `\`) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") {
			return fmt.Errorf("'source.path' must be a clean relative path, got %q", p)
		}
	}
	if r := s.Revision; r != "" {
		if len(r) < 7 || len(r) > 64 || strings.Trim(r, "0123456789abcdef") != "" {
			return fmt.Errorf("'source.revision' must be a lowercase hex commit id, got %q", r)
		}
	}
	return nil
}

// URL returns a browsable link to the module's source. For the common forges
// it points at the directory at the declared revision (or the default branch);
// for any other host it is the repository URL itself.
func (s Source) URL() string {
	if s.Repository == "" {
		return ""
	}
	repo := strings.TrimSuffix(s.Repository, "/")
	if s.Path == "" && s.Revision == "" {
		return repo
	}
	rev := s.Revision
	if rev == "" {
		rev = "HEAD"
	}
	u, err := url.Parse(repo)
	if err != nil {
		return repo
	}
	var link string
	switch u.Host {
	case "github.com":
		link = repo + "/tree/" + rev
	case "gitlab.com":
		link = repo + "/-/tree/" + rev
	default:
		return repo
	}
	if s.Path != "" {
		link += "/" + s.Path
	}
	return link
}

type Replace struct {
	// Old is the old module path of the dependency, i.e.,
	// `gno.land/r/path/to/module`.
	Old string `toml:"old" json:"old"`
	// New is the new module path of the dependency, i.e.,
	// `gno.land/r/path/to/module/v2` or a local path, i.e.,
	// `../path/to/module`.
	New string `toml:"new" json:"new"`
}

// GetGno returns the current gno version or the default one.
func (f *File) GetGno() (version string) {
	if f.Gno == "" {
		return "0.0"
	}
	return f.Gno
}

// SetGno sets the gno version.
func (f *File) SetGno(version string) {
	f.Gno = version
}

// AddReplace adds a replace directive or replaces an existing one.
func (f *File) AddReplace(oldPath, newPath string) {
	for i, r := range f.Replace {
		if r.Old == oldPath {
			f.Replace[i].New = newPath
			return
		}
	}
	newReplace := Replace{Old: oldPath, New: newPath}
	f.Replace = append(f.Replace, newReplace)
}

// DropReplace drops a replace directive.
func (f *File) DropReplace(oldPath string) {
	for i, r := range f.Replace {
		if r.Old == oldPath {
			f.Replace = append(f.Replace[:i], f.Replace[i+1:]...)
		}
	}
}

// Validate validates gnomod.toml.
func (f *File) Validate() error {
	modPath := f.Module

	// module is required.
	if modPath == "" {
		return fmt.Errorf("invalid gnomod.toml: 'module' is required")
	}

	// module is a valid import path.
	err := module.CheckImportPath(modPath)
	if err != nil {
		return fmt.Errorf("invalid gnomod.toml: %w", err)
	}

	if err := f.Source.Validate(); err != nil {
		return fmt.Errorf("invalid gnomod.toml: %w", err)
	}

	return nil
}

// Resolve takes a module path and returns any adequate replacement following
// the Replace directives.
func (f *File) Resolve(target string) string {
	for _, r := range f.Replace {
		if r.Old == target {
			return r.New
		}
	}
	return target
}

// WriteFile writes gnomod.toml to the given absolute file path.
func (f *File) WriteFile(fpath string) error {
	data := []byte(f.WriteString())
	err := os.WriteFile(fpath, data, 0o644)
	if err != nil {
		return fmt.Errorf("writefile %q: %w", fpath, err)
	}
	return nil
}

// Sanitize sanitizes the gnomod.toml file.
func (f *File) Sanitize() {
	// set default version if missing.
	f.Gno = f.GetGno()

	// sanitize replaces.
	replaces := make([]Replace, 0, len(f.Replace))
	seen := make(map[string]bool)
	for _, r := range f.Replace {
		// empty replaces.
		if r.Old == "" || r.New == "" || r.Old == r.New {
			continue
		}

		// duplicates.
		if seen[r.Old] {
			continue
		}
		seen[r.Old] = true

		replaces = append(replaces, r)
	}
	f.Replace = replaces
}

// HasReplaces returns true if the module has any replace directives.
func (f *File) HasReplaces() bool {
	return len(f.Replace) > 0
}
