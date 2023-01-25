package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/std"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var (
	source        = "test3.gno.land:36657" // "127.0.0.1:26657"
	queryPathFile = "vm/qfile"
)

// ReadModFile reads, parses and validates the mod file at gnomod.
func ReadModFile(absModPath string) (f *File, err error) {
	data, err := os.ReadFile(absModPath)
	if err != nil {
		return nil, fmt.Errorf("reading gno.mod: %s", err)
	}

	f, err = Parse(absModPath, data)
	if err != nil {
		return nil, fmt.Errorf("parsing gno.mod: %s", err)
	}

	return f, err
}

func IsModFileExist(absModPath string) bool {
	_, err := os.Stat(absModPath)
	return err == nil
}

// GetGnoModPath returns the path for gno modules
func GetGnoModPath() (string, error) {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return "", errors.New("GOPATH not found")
	}

	return filepath.Join(goPath, "pkg/gnomod"), nil
}

// FetchModPackages fetches and writes gno.mod packages
// in GOPATH/pkg/gnomod/
func FetchModPackages(f *File) error {
	gnoModPath, err := GetGnoModPath()
	if err != nil {
		return fmt.Errorf("fetching mods: %s", err)
	}

	if f.Require != nil {
		for _, r := range f.Require {
			fmt.Println("fetching", r.Mod.Path)
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

			WriteGoMod(filepath.Join(gnoModPath, r.Mod.Path), f)
		}
	}

	return nil
}

func writePackage(basePath, pkgPath string) error {
	res, err := QueryChain(queryPathFile, []byte(pkgPath))
	if err != nil {
		return fmt.Errorf("makeReq gno.mod: %s", err)
	}

	dirPath, fileName := std.SplitFilepath(pkgPath)
	if fileName == "" {
		// Is Dir
		// Create Dir if not exists
		dirPath := filepath.Join(basePath, dirPath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := os.MkdirAll(dirPath, 0o755); err != nil {
				return fmt.Errorf("creating pkg dir: %s", err)
			}
		}

		files := strings.Split(string(res.Data), "\n")
		for _, file := range files {
			if err := writePackage(basePath, filepath.Join(pkgPath, file)); err != nil {
				return fmt.Errorf("writing mod files: %s", err)
			}
		}
	} else {
		// Is File
		err := os.WriteFile(filepath.Join(basePath, pkgPath), []byte(res.Data), 0o644)
		if err != nil {
			return fmt.Errorf("writing mod files: %s", err)
		}
	}

	return nil
}

// WriteGoMod writes go.mod file in the given absolute path
// TODO: Find better way to do this.
func WriteGoMod(absPath string, f *File) error {
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

// ReplaceModuleAll replaces all the required modules with
// the modules in given path.
func ReplaceModuleAll(f *File, path string) {
	for _, req := range f.Require {
		f.Replace = append(f.Replace, &modfile.Replace{
			Old: module.Version{
				Path:    req.Mod.Path,
				Version: req.Mod.Version,
			},
			New: module.Version{
				Path: filepath.Join(path, req.Mod.Path),
			},
		})
	}
}
