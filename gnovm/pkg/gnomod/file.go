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
	Go      *modfile.Go
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}

func (f *File) AddModuleStmt(path string) error {
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
	return nil
}

func (f *File) AddComment(text string) {
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

func (f *File) AddReplace(oldPath, oldVers, newPath, newVers string) error {
	return addReplace(f.Syntax, &f.Replace, oldPath, oldVers, newPath, newVers)
}

func (f *File) DropReplace(oldPath, oldVers string) error {
	for _, r := range f.Replace {
		if r.Old.Path == oldPath && r.Old.Version == oldVers {
			markLineAsRemoved(r.Syntax)
			*r = modfile.Replace{}
		}
	}
	return nil
}

// Validate validates gno.mod
func (f *File) Validate() error {
	if f.Module == nil {
		return errors.New("requires module")
	}

	return nil
}

// Resolve takes a Require directive from File and returns any adequate replacement
// following the Replace directives.
func (f *File) Resolve(r *modfile.Require) module.Version {
	mod, replaced := isReplaced(r.Mod, f.Replace)
	if replaced {
		return mod
	}
	return r.Mod
}

// writes file to the given absolute file path
func (f *File) Write(fname string) error {
	f.Syntax.Cleanup()
	data := modfile.Format(f.Syntax)
	err := os.WriteFile(fname, data, 0o644)
	if err != nil {
		return fmt.Errorf("writefile %q: %w", fname, err)
	}
	return nil
}

func (f *File) Sanitize() {
	removeDups(f.Syntax, &f.Replace)
}
