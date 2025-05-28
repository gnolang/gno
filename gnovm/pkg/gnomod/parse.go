package gnomod

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var (
	ErrNoModFile  = errors.New("gno.mod doesn't exist")
	ErrNoTomlFile = errors.New("gnomod.toml doesn't exist")
)

// ParseDir parses, validates and returns a gno.mod or gnomod.toml file located at dir or at
// dir's parents.
func ParseDir(dir string) (*File, error) {
	ferr := func(err error) (*File, error) {
		return nil, fmt.Errorf("parsing gno.mod/gnomod.toml at %s: %w", dir, err)
	}

	// FindRootDir requires absolute path, make sure its the case
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return ferr(err)
	}
	rd, err := FindRootDir(absDir)
	if err != nil {
		return ferr(err)
	}

	// Try gnomod.toml first
	tomlPath := filepath.Join(rd, "gnomod.toml")
	if _, err := os.Stat(tomlPath); err == nil {
		return ParseTomlFile(tomlPath)
	}

	// Fall back to gno.mod
	modPath := filepath.Join(rd, "gno.mod")
	if _, err := os.Stat(modPath); err != nil {
		return ferr(err)
	}
	b, err := os.ReadFile(modPath)
	if err != nil {
		return ferr(err)
	}
	dmf, err := ParseModBytes(modPath, b)
	if err != nil {
		return ferr(err)
	}
	if err := dmf.Validate(); err != nil {
		return ferr(err)
	}
	return FromDeprecatedModFile(dmf), nil
}

// ParseFilepath tries to parse gno.mod or gnomod.toml file given the file path
func ParseFilepath(fpath string) (*File, error) {
	file, err := os.Stat(fpath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}
	if file.IsDir() {
		return nil, fmt.Errorf("invalid file at %q: is a directory", fpath)
	}

	// Check if it's a TOML file
	if filepath.Ext(fpath) == ".toml" {
		return ParseTomlFile(fpath)
	}

	// Assume it's a gno.mod file
	b, err := os.ReadFile(fpath)
	if err != nil {
		return nil, fmt.Errorf("could not read gno.mod file: %w", err)
	}
	dmf, err := ParseModBytes(fpath, b)
	if err != nil {
		return nil, fmt.Errorf("error parsing gno.mod file at %q: %w", fpath, err)
	}
	if err := dmf.Validate(); err != nil {
		return nil, fmt.Errorf("error validating gno.mod file at %q: %w", fpath, err)
	}
	return FromDeprecatedModFile(dmf), nil
}

// ParseModBytes parses and returns a gno.mod file.
//
// - fname is the name of the file, used in positions and errors.
// - data is the content of the file.
func ParseModBytes(fname string, data []byte) (*DeprecatedModFile, error) {
	fs, err := parse(fname, data)
	if err != nil {
		return nil, err
	}
	f := &DeprecatedModFile{
		Syntax: fs,
	}
	var errs modfile.ErrorList

	for _, x := range fs.Stmt {
		switch x := x.(type) {
		case *modfile.Line:
			f.add(&errs, nil, x, x.Token[0], x.Token[1:])
		case *modfile.LineBlock:
			if len(x.Token) > 1 {
				errs = append(errs, modfile.Error{
					Filename: fname,
					Pos:      x.Start,
					Err:      fmt.Errorf("unknown block type: %s", strings.Join(x.Token, " ")),
				})
				continue
			}
			switch x.Token[0] {
			default:
				errs = append(errs, modfile.Error{
					Filename: fname,
					Pos:      x.Start,
					Err:      fmt.Errorf("unknown block type: %s", strings.Join(x.Token, " ")),
				})
				continue
			case "module", "replace":
				for _, l := range x.Line {
					f.add(&errs, x, l, x.Token[0], l.Token)
				}
			}
		case *modfile.CommentBlock:
			if x.Start.Line == 1 {
				f.Draft = parseDraft(x)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}
	return f, nil
}

// ParseMemPackage parses gno.mod or gnomod.toml from MemPackage
func ParseMemPackage(mpkg *std.MemPackage) (*File, error) {
	// Try gnomod.toml first
	if mf := mpkg.GetFile("gnomod.toml"); mf != nil {
		return ParseTomlFile(mf.Name)
	}

	// Fall back to gno.mod
	mf := mpkg.GetFile("gno.mod")
	if mf == nil {
		return nil, fmt.Errorf(
			"gno.mod not in mem package %s (name=%s): %w",
			mpkg.Path, mpkg.Name, os.ErrNotExist,
		)
	}
	dmf, err := ParseModBytes(mf.Name, []byte(mf.Body))
	if err != nil {
		return nil, err
	}
	return FromDeprecatedModFile(dmf), nil
}

// MustParseMemPackage parses gno.mod or gnomod.toml from MemPackage, panicking on error
func MustParseMemPackage(mpkg *std.MemPackage) *File {
	mod, err := ParseMemPackage(mpkg)
	if err != nil {
		panic(fmt.Errorf("parsing mempackage %w", err))
	}
	return mod
}

var reGnoVersion = regexp.MustCompile(`^([0-9][0-9]*)\.(0|[1-9][0-9]*)(\.(0|[1-9][0-9]*))?([a-z]+[0-9]+)?$`)

func (f *DeprecatedModFile) add(errs *modfile.ErrorList, block *modfile.LineBlock, line *modfile.Line, verb string, args []string) {
	wrapError := func(err error) {
		*errs = append(*errs, modfile.Error{
			Filename: f.Syntax.Name,
			Pos:      line.Start,
			Err:      err,
		})
	}
	errorf := func(format string, args ...any) {
		wrapError(fmt.Errorf(format, args...))
	}

	switch verb {
	default:
		errorf("unknown directive: %s", verb)

	case "gno":
		if f.Gno != nil {
			errorf("repeated gno statement")
			return
		}
		if len(args) != 1 {
			errorf("gno directive expects exactly one argument")
			return
		} else if !reGnoVersion.MatchString(args[0]) {
			fixed := false
			if !fixed {
				errorf("invalid gno version %s: must match format 1.23", args[0])
				return
			}
		}

		line := reflect.ValueOf(line).Interface().(*modfile.Line)
		f.Gno = &modfile.Go{Syntax: line}
		f.Gno.Version = args[0]

	case "module":
		if f.Module != nil {
			errorf("repeated module statement")
			return
		}
		deprecated := parseDeprecation(block, line)
		f.Module = &modfile.Module{
			Syntax:     line,
			Deprecated: deprecated,
		}
		if len(args) != 1 {
			errorf("usage: module module/path")
			return
		}
		s, err := parseString(&args[0])
		if err != nil {
			errorf("invalid quoted string: %v", err)
			return
		}
		if err := module.CheckImportPath(s); err != nil {
			errorf("invalid module path: %v", err)
			return
		}
		f.Module.Mod = module.Version{Path: s}

	case "replace":
		replace, wrappederr := parseReplace(f.Syntax.Name, line, verb, args)
		if wrappederr != nil {
			*errs = append(*errs, *wrappederr)
			return
		}
		f.Replace = append(f.Replace, replace)
	}
}

// Validate validates gno.mod
func (f *DeprecatedModFile) Validate() error {
	if f.Module == nil {
		return errors.New("requires module")
	}
	return nil
}
