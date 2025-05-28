// Some part of file is copied and modified from
// golang.org/x/mod/modfile/read.go
//
// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in here[1].
//
// [1]: https://cs.opensource.google/go/x/mod/+/master:LICENSE

package gnomod

import (
	"errors"
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
	"golang.org/x/mod/module"
)

// Parsed gno.mod file.
type File struct {
	// Module is the module section of the gno.mod file.
	// It is intended to be the main place for manual customization by the
	// author of the module.
	Module struct {
		// Path is the path of the module.
		// Like `gno.land/r/path/to/module`.
		Path string `toml:"path" json:"path"`

		// Draft indicates that the module isn't ready for production use.
		//
		// Draft modules:
		// - are added to the chain at genesis time and cannot be added after.
		// - cannot be imported by other newly added modules.
		Draft bool `toml:"draft" json:"draft"`

		// Private indicates that the module is private.
		//
		// Private modules:
		// - cannot be imported by other modules.
		// -
		Private bool `toml:"private" json:"private"`

		// XXX: Version // version of the module?
	} `toml:"module" json:"module"`

	// Develop is the develop section of the gno.mod file.
	//
	// It is wiped out by the vmkeeper when a module is added.
	Develop struct {
		// Replace allows specifying a replacement for a module.
		//
		// It can link to a different online module path, or a local path.
		// If this value is set, the module cannot be added to the chain.
		Replace []Replace `toml:"replace" json:"replace"`
	} `toml:"develop" json:"develop"`

	// Gno is the gno section of the gno.mod file.
	//
	// It is used to specify the compatibility of the module within the gno
	// toolchain.
	// It is intended to be set by the `gno` cli when initializing or upgrading
	// a module.
	Gno struct {
		Version string `toml:"version" json:"version"`
	} `toml:"gno" json:"gno"`

	// UploadMetadata is the upload metadata section of the gno.mod file.
	//
	// Is it filled by the vmkeeper when a module is added.
	// It is not intended to be used offchain.
	UploadMetadata struct {
		Uploader   string `toml:"uploader" json:"uploader"`
		UploadedAt string `toml:"uploaded_at" json:"uploaded_at"`
		// XXX: GnoVersion // gno version at upload time?
	} `toml:"upload_metadata" json:"upload_metadata"`
}

// Replace is a replace directive for one of the module's dependencies.
type Replace struct {
	// Old is the old module path of the dependency, i.e.,
	// `gno.land/r/path/to/module`.
	Old string `toml:"old" json:"old"`

	// New is the new module path of the dependency, i.e.,
	// `gno.land/r/path/to/module/v2` or a local path, i.e.,
	// `../path/to/module`.
	New string `toml:"new" json:"new"`
}

// GetGnoVersion returns the current gno version or the default one.
func (f *File) GetGnoVersion() (version string) {
	if f.Gno.Version == "" {
		return "0.0"
	}
	return f.Gno.Version
}

// AddReplace adds a replace directive or replaces an existing one.
func (f *File) AddReplace(oldPath, newPath string) {
	for _, r := range f.Develop.Replace {
		if r.Old == oldPath {
			r.New = newPath
			return
		}
	}
	newReplace := Replace{Old: oldPath, New: newPath}
	f.Develop.Replace = append(f.Develop.Replace, newReplace)
}

// DropReplace drops a replace directive.
func (f *File) DropReplace(oldPath string) {
	for i, r := range f.Develop.Replace {
		if r.Old == oldPath {
			f.Develop.Replace = append(f.Develop.Replace[:i], f.Develop.Replace[i+1:]...)
		}
	}
}

// Validate validates gnomod.toml.
func (f *File) Validate() error {
	modPath := f.Module.Path

	// module.path is required.
	if modPath == "" {
		return errors.New("requires module.path")
	}

	// module.path is a valid import path.
	err := module.CheckImportPath(modPath)
	if err != nil {
		return err
	}

	return nil
}

// Resolve takes a module path and returns any adequate replacement following
// the Replace directives.
func (f *File) Resolve(target string) string {
	for _, r := range f.Develop.Replace {
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

// writes to a string.
func (f *File) WriteString() string {
	data, err := toml.Marshal(f)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// Sanitize sanitizes the gnomod.toml file.
func (f *File) Sanitize() {
	// set default version if missing.
	f.Gno.Version = f.GetGnoVersion()

	// sanitize develop.replaces.
	replaces := make([]Replace, 0, len(f.Develop.Replace))
	seen := make(map[string]bool)
	for _, r := range f.Develop.Replace {
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
	f.Develop.Replace = replaces
}
