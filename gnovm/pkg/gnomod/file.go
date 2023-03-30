package gnomod

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
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

// Validate validates gno.mod
func (f *File) Validate() error {
	if f.Module == nil {
		return errors.New("requires module")
	}

	return nil
}

// FetchDeps fetches and writes gno.mod packages
// in GOPATH/pkg/gnomod/
func (f *File) FetchDeps(path string, remote string) error {
	for _, r := range f.Require {
		mod, replaced := isReplaced(r.Mod, f.Replace)
		if replaced {
			if modfile.IsDirectoryPath(mod.Path) {
				continue
			}
			r.Mod = *mod
		}
		indirect := ""
		if r.Indirect {
			indirect = "// indirect"
		}

		_, err := os.Stat(filepath.Join(path, r.Mod.Path))
		if !os.IsNotExist(err) {
			log.Println("cached", r.Mod.Path, indirect)
			continue
		}
		log.Println("fetching", r.Mod.Path, indirect)
		requirements, err := writePackage(remote, path, r.Mod.Path)
		if err != nil {
			return fmt.Errorf("writepackage: %w", err)
		}

		modFile := &File{
			Module: &modfile.Module{
				Mod: module.Version{
					Path: r.Mod.Path,
				},
			},
		}
		for _, req := range requirements {
			path := req[1 : len(req)-1] // trim leading and trailing `"`
			if strings.HasSuffix(path, modFile.Module.Mod.Path) {
				continue
			}
			// skip if `std`, special case.
			if path == gnolang.GnoStdPkgAfter {
				continue
			}

			if strings.HasPrefix(path, gnolang.ImportPrefix) {
				path = strings.TrimPrefix(path, gnolang.ImportPrefix+"/examples/")
				modFile.Require = append(modFile.Require, &modfile.Require{
					Mod: module.Version{
						Path:    path,
						Version: "v0.0.0", // TODO: Use latest?
					},
					Indirect: true,
				})
			}
		}

		err = modFile.FetchDeps(path, remote)
		if err != nil {
			return err
		}
		goMod, err := GnoToGoMod(*modFile)
		if err != nil {
			return err
		}
		err = goMod.WriteToPath(filepath.Join(path, r.Mod.Path))
		if err != nil {
			return err
		}
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

	modPath := filepath.Join(absPath, "go.mod")
	err := os.WriteFile(modPath, []byte(data), 0o644)
	if err != nil {
		return fmt.Errorf("writefile %q: %w", modPath, err)
	}

	return nil
}

func (f *File) Sanitize() {
	removeDups(f.Syntax, &f.Require, &f.Replace)
}
