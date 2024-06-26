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
	"log"
	"os"
	"path/filepath"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Parsed gno.mod file.
type File struct {
	Draft   bool
	Module  *modfile.Module
	Go      *modfile.Go
	Require []*modfile.Require
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}

// AddRequire sets the first require line for path to version vers,
// preserving any existing comments for that line and removing all
// other lines for path.
//
// If no line currently exists for path, AddRequire adds a new line
// at the end of the last require block.
func (f *File) AddRequire(path, vers string) error {
	need := true
	for _, r := range f.Require {
		if r.Mod.Path == path {
			if need {
				r.Mod.Version = vers
				updateLine(r.Syntax, "require", modfile.AutoQuote(path), vers)
				need = false
			} else {
				markLineAsRemoved(r.Syntax)
				*r = modfile.Require{}
			}
		}
	}

	if need {
		f.AddNewRequire(path, vers, false)
	}
	return nil
}

// AddNewRequire adds a new require line for path at version vers at the end of
// the last require block, regardless of any existing require lines for path.
func (f *File) AddNewRequire(path, vers string, indirect bool) {
	line := addLine(f.Syntax, nil, "require", modfile.AutoQuote(path), vers)
	r := &modfile.Require{
		Mod:    module.Version{Path: path, Version: vers},
		Syntax: line,
	}
	setIndirect(r, indirect)
	f.Require = append(f.Require, r)
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

func (f *File) DropRequire(path string) error {
	for _, r := range f.Require {
		if r.Mod.Path == path {
			markLineAsRemoved(r.Syntax)
			*r = modfile.Require{}
		}
	}
	return nil
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

// FetchDeps fetches and writes gno.mod packages
// in GOPATH/pkg/gnomod/
func (f *File) FetchDeps(path string, remote string, verbose bool) error {
	for _, r := range f.Require {
		mod := f.Resolve(r)
		if r.Mod.Path != mod.Path {
			if modfile.IsDirectoryPath(mod.Path) {
				continue
			}
		}
		indirect := ""
		if r.Indirect {
			indirect = "// indirect"
		}

		_, err := os.Stat(PackageDir(path, mod))
		if !os.IsNotExist(err) {
			if verbose {
				log.Println("cached", mod.Path, indirect)
			}
			continue
		}
		if verbose {
			log.Println("fetching", mod.Path, indirect)
		}
		requirements, err := writePackage(remote, path, mod.Path)
		if err != nil {
			return fmt.Errorf("writepackage: %w", err)
		}

		modFile := new(File)
		modFile.AddModuleStmt(mod.Path)
		for _, req := range requirements {
			path := req[1 : len(req)-1] // trim leading and trailing `"`
			if strings.HasSuffix(path, modFile.Module.Mod.Path) {
				continue
			}

			if !gno.IsStdlib(path) {
				modFile.AddNewRequire(path, "v0.0.0-latest", true)
			}
		}

		err = modFile.FetchDeps(path, remote, verbose)
		if err != nil {
			return err
		}
		goMod, err := GnoToGoMod(*modFile)
		if err != nil {
			return err
		}
		pkgPath := PackageDir(path, mod)
		goModFilePath := filepath.Join(pkgPath, "go.mod")
		err = goMod.Write(goModFilePath)
		if err != nil {
			return err
		}
	}

	return nil
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
	removeDups(f.Syntax, &f.Require, &f.Replace)
}
