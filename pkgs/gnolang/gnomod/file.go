package gnomod

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// Parsed gno.mod file.
type File struct {
	Module  *modfile.Module
	Go      *modfile.Go
	Require []*modfile.Require
	Replace []*modfile.Replace

	Syntax *modfile.FileSyntax
}

// FetchDeps fetches and writes gno.mod packages
// in GOPATH/pkg/gnomod/
func (f *File) FetchDeps() error {
	gnoModPath, err := GetGnoModPath()
	if err != nil {
		return fmt.Errorf("fetching mods: %s", err)
	}

	for _, r := range f.Require {
		log.Println("fetching", r.Mod.Path)
		err := writePackage(gnoModPath, r.Mod.Path)
		if err != nil {
			return fmt.Errorf("fetching mods: %s", err)
		}

		f := &File{
			Module: &modfile.Module{
				Mod: module.Version{
					Path: r.Mod.Path,
				},
			},
		}

		f.WriteToPath(filepath.Join(gnoModPath, r.Mod.Path))
	}

	return nil
}

// WriteToPath writes go.mod file in the given absolute path
// TODO: Find better way to do this. Try to use `modfile`
// package to manage this.
func (f *File) WriteToPath(absPath string) error {
	if f.Module == nil {
		return errors.New("writing go.mod: module not found")
	}

	data := "module " + f.Module.Mod.Path + "\n"

	if f.Go != nil {
		data += "\ngo " + f.Go.Version + "\n"
	}

	if f.Require != nil {
		data += "\nrequire (" + "\n"
		for _, req := range f.Require {
			data += "\t" + req.Mod.Path + " " + req.Mod.Version + "\n"
		}
		data += ")\n"
	}

	if f.Replace != nil {
		data += "\nreplace (" + "\n"
		for _, rep := range f.Replace {
			data += "\t" + rep.Old.Path + " " + rep.Old.Version +
				" => " + rep.New.Path + "\n"
		}
		data += ")\n"
	}

	err := os.WriteFile(filepath.Join(absPath, "go.mod"), []byte(data), 0o644)
	if err != nil {
		return fmt.Errorf("writing go.mod: %s", err)
	}

	return nil
}
