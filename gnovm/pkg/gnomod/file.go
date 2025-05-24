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

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Parsed gno.mod file.
type File struct {
	Draft   bool
	Module  *modfile.Module
	Gno     *modfile.Go
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}

func (f *File) GetGno() (version string) {
	if f.Gno == nil {
		return "0.0"
	} else {
		return f.Gno.Version
	}
}

func (f *File) SetGno(version string) {
	if f.Syntax == nil {
		f.Syntax = new(modfile.FileSyntax)
	}
	if f.Gno == nil {
		f.Gno = &modfile.Go{
			Version: version,
			Syntax:  addLine(f.Syntax, nil, "gno", version),
		}
	} else {
		f.Gno.Version = version
		updateLine(f.Gno.Syntax, "gno", version)
	}
}

func (f *File) SetModule(path string) {
	if f.Syntax == nil {
		f.Syntax = new(modfile.FileSyntax)
	}
	if f.Module == nil {
		f.Module = &modfile.Module{
			Mod:    module.Version{Path: path},
			Syntax: addLine(f.Syntax, nil, "module", modfile.AutoQuote(path)),
		}
	} else {
		f.Module.Mod.Path = path
		updateLine(f.Module.Syntax, "module", modfile.AutoQuote(path))
	}
}

func (f *File) SetComment(text string) {
	if f.Syntax == nil {
		f.Syntax = new(modfile.FileSyntax)
	}
	f.Syntax.Stmt = append(f.Syntax.Stmt, &modfile.CommentBlock{
		Comments: modfile.Comments{
			Before: []modfile.Comment{
				{
					Token: text,
				},
			},
		},
	})
}

func (f *File) AddReplace(oldPath, oldVers, newPath, newVers string) {
	addReplace(f.Syntax, &f.Replace, oldPath, oldVers, newPath, newVers)
}

func (f *File) DropReplace(oldPath, oldVers string) {
	for _, r := range f.Replace {
		if r.Old.Path == oldPath && r.Old.Version == oldVers {
			markLineAsRemoved(r.Syntax)
			*r = modfile.Replace{}
		}
	}
}

// Validate validates gno.mod
func (f *File) Validate() error {
	if f.Module == nil {
		return errors.New("requires module")
	}

	return nil
}

// Resolve takes a module version and returns any adequate replacement
// following the Replace directives.
func (f *File) Resolve(m module.Version) module.Version {
	if f == nil {
		return m
	}
	mod, replaced := isReplaced(m, f.Replace)
	if replaced {
		return mod
	}
	return m
}

// writes file to the given absolute file path
func (f *File) WriteFile(fpath string) error {
	f.Syntax.Cleanup()
	data := modfile.Format(f.Syntax)
	err := os.WriteFile(fpath, data, 0o644)
	if err != nil {
		return fmt.Errorf("writefile %q: %w", fpath, err)
	}
	return nil
}

// writes to a string
func (f *File) WriteString() string {
	f.Syntax.Cleanup()
	data := modfile.Format(f.Syntax)
	return string(data)
}

func (f *File) Sanitize() {
	removeDups(f.Syntax, &f.Replace)
}
