package gnomod

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

// ParseAt parses, validates and returns a gno.mod file located at dir or at
// dir's parents.
func ParseAt(dir string) (*File, error) {
	ferr := func(err error) (*File, error) {
		return nil, fmt.Errorf("parsing gno.mod at %s: %w", dir, err)
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
	fname := filepath.Join(rd, "gno.mod")
	b, err := os.ReadFile(fname)
	if err != nil {
		return ferr(err)
	}
	gm, err := Parse(fname, b)
	if err != nil {
		return ferr(err)
	}
	if err := gm.Validate(); err != nil {
		return ferr(err)
	}
	return gm, nil
}

// tries to parse gno mod file given the filename, using Parse and Validate from
// the gnomod package
//
// TODO(tb): replace by `gnomod.ParseAt` ? The key difference is the latter
// looks for gno.mod in parent directories, while this function doesn't.
func ParseGnoMod(fname string) (*File, error) {
	file, err := os.Stat(fname)
	if err != nil {
		return nil, fmt.Errorf("could not read gno.mod file: %w", err)
	}
	if file.IsDir() {
		return nil, fmt.Errorf("invalid gno.mod at %q: is a directory", fname)
	}

	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("could not read gno.mod file: %w", err)
	}
	gm, err := Parse(fname, b)
	if err != nil {
		return nil, fmt.Errorf("error parsing gno.mod file at %q: %w", fname, err)
	}
	if err := gm.Validate(); err != nil {
		return nil, fmt.Errorf("error validating gno.mod file at %q: %w", fname, err)
	}
	return gm, nil
}

// Parse parses and returns a gno.mod file.
//
// - file is the name of the file, used in positions and errors.
// - data is the content of the file.
func Parse(file string, data []byte) (*File, error) {
	fs, err := parse(file, data)
	if err != nil {
		return nil, err
	}
	f := &File{
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
					Filename: file,
					Pos:      x.Start,
					Err:      fmt.Errorf("unknown block type: %s", strings.Join(x.Token, " ")),
				})
				continue
			}
			switch x.Token[0] {
			default:
				errs = append(errs, modfile.Error{
					Filename: file,
					Pos:      x.Start,
					Err:      fmt.Errorf("unknown block type: %s", strings.Join(x.Token, " ")),
				})
				continue
			case "module", "require", "replace":
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

func (f *File) add(errs *modfile.ErrorList, block *modfile.LineBlock, line *modfile.Line, verb string, args []string) {
	wrapError := func(err error) {
		*errs = append(*errs, modfile.Error{
			Filename: f.Syntax.Name,
			Pos:      line.Start,
			Err:      err,
		})
	}
	errorf := func(format string, args ...interface{}) {
		wrapError(fmt.Errorf(format, args...))
	}

	switch verb {
	default:
		errorf("unknown directive: %s", verb)

	case "go":
		if f.Go != nil {
			errorf("repeated go statement")
			return
		}
		if len(args) != 1 {
			errorf("go directive expects exactly one argument")
			return
		} else if !modfile.GoVersionRE.MatchString(args[0]) {
			fixed := false
			if !fixed {
				errorf("invalid go version '%s': must match format 1.23", args[0])
				return
			}
		}

		line := reflect.ValueOf(line).Interface().(*modfile.Line)
		f.Go = &modfile.Go{Syntax: line}
		f.Go.Version = args[0]

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
		f.Module.Mod = module.Version{Path: s}

	case "require":
		if len(args) != 2 {
			errorf("usage: %s module/path v1.2.3", verb)
			return
		}
		s, err := parseString(&args[0])
		if err != nil {
			errorf("invalid quoted string: %v", err)
			return
		}
		v, err := parseVersion(verb, s, &args[1])
		if err != nil {
			wrapError(err)
			return
		}
		f.Require = append(f.Require, &modfile.Require{
			Mod:    module.Version{Path: s, Version: v},
			Syntax: line,
		})

	case "replace":
		replace, wrappederr := parseReplace(f.Syntax.Name, line, verb, args)
		if wrappederr != nil {
			*errs = append(*errs, *wrappederr)
			return
		}
		f.Replace = append(f.Replace, replace)
	}
}
