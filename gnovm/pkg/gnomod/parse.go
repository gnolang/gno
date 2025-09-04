package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/std"
)

var ErrNoModFile = errors.New("gnomod.toml doesn't exist")

// ParseDir parses, validates and returns a gno.mod or gnomod.toml file located at dir (does not search parents).
func ParseDir(dir string) (*File, error) {
	ferr := func(err error) (*File, error) {
		return nil, fmt.Errorf("parsing gno.mod/gnomod.toml at %s: %w", dir, err)
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ferr(err)
	}

	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		fpath := filepath.Join(absDir, fname)
		if _, err := os.Stat(fpath); err == nil {
			b, err := os.ReadFile(fpath)
			if err != nil {
				return ferr(err)
			}
			return ParseBytes(fpath, b)
		}
	}

	return ferr(ErrNoModFile)
}

// ParseFilepath tries to parse gno.mod or gnomod.toml file given the file path.
func ParseFilepath(fpath string) (*File, error) {
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("could not read file %q: %w", fpath, err)
	}
	return ParseBytes(fpath, b)
}

// MustParseBytes parses a gnomod.toml or gno.mod file from bytes or panic.
func MustParseBytes(fname string, data []byte) *File {
	mod, err := ParseBytes(fname, data)
	if err != nil {
		panic(fmt.Errorf("parsing bytes %w", err))
	}
	return mod
}

// ParseBytes parses a gnomod.toml or gno.mod file from bytes.
func ParseBytes(fpath string, data []byte) (*File, error) {
	f, err := parseBytes(fpath, data)
	if err != nil {
		return nil, err
	}
	if err := f.Validate(); err != nil {
		return nil, err
	}
	return f, nil
}

func parseBytes(fpath string, data []byte) (*File, error) {
	fname := filepath.Base(fpath)

	// gnomod.toml
	switch fname {
	case "gnomod.toml":
		return parseTomlBytes(fname, data)
	case "gno.mod":
		dmf, err := parseDeprecatedDotModBytes(fname, data)
		if err != nil {
			return nil, err
		}
		return dmf.Migrate()
	}

	return nil, fmt.Errorf("invalid file at %q: unknown file type", fpath)
}

// ParseMemPackage parses gnomod.toml or gno.mod from MemPackage.
func ParseMemPackage(mpkg *std.MemPackage) (*File, error) {
	for _, fname := range []string{"gnomod.toml", "gno.mod"} {
		if mf := mpkg.GetFile(fname); mf != nil {
			return ParseBytes(mf.Name, []byte(mf.Body))
		}
	}
	return nil, fmt.Errorf("gnomod.toml not in mem package %s (name=%s): %w", mpkg.Path, mpkg.Name, os.ErrNotExist)
}

// MustParseMemPackage parses gno.mod or gnomod.toml from MemPackage,
// panicking on error.
func MustParseMemPackage(mpkg *std.MemPackage) *File {
	mod, err := ParseMemPackage(mpkg)
	if err != nil {
		panic(fmt.Errorf("parsing mempackage: %w", err))
	}
	return mod
}

var reGnoVersion = regexp.MustCompile(`^([0-9][0-9]*)\.(0|[1-9][0-9]*)(\.(0|[1-9][0-9]*))?([a-z]+[0-9]+)?$`)
