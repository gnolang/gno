package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/std"
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

func FetchModPackages(f *File) error {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		return errors.New("GOPATH not found")
	}

	if f.Require != nil {
		for _, r := range f.Require {
			fmt.Println("fetching", r.Mod.Path)
			if err := writeModFiles(filepath.Join(goPath, "pkg/gnomod"), r.Mod.Path); err != nil {
				return fmt.Errorf("fetching mods: %s", err)
			}
		}
	}

	// TODO: Remove
	//fmt.Println(f.Replace[0].New.Path)

	return nil
}

// writeModFiles recursively fetches and writes gno modules
// in {basePath}/{pkgPath}
func writeModFiles(basePath, pkgPath string) error {
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
			if err := writeModFiles(basePath, filepath.Join(pkgPath, file)); err != nil {
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
