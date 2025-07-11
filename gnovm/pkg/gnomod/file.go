package gnomod

import (
	"fmt"
	"os"

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
	// XXX: GnoVersion // gno version at add time?
	// XXX: Consider things like IsUsingBanker or other security-awareness flags
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
