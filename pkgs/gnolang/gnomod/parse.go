package gnomod

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

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
		f.Module = &modfile.Module{
			Syntax: line,
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
		if len(args) != 1 {
			errorf("usage: %s module/path", verb)
			return
		}
		s, err := parseString(&args[0])
		if err != nil {
			errorf("invalid quoted string: %v", err)
			return
		}
		f.Require = append(f.Require, &modfile.Require{
			Mod:    module.Version{Path: s, Version: "v0.0.0"},
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

func parseReplace(filename string, line *modfile.Line, verb string, args []string) (*modfile.Replace, *modfile.Error) {
	wrapError := func(err error) *modfile.Error {
		return &modfile.Error{
			Filename: filename,
			Pos:      line.Start,
			Err:      err,
		}
	}
	errorf := func(format string, args ...interface{}) *modfile.Error {
		return wrapError(fmt.Errorf(format, args...))
	}

	if len(args) != 3 || args[1] != "=>" {
		return nil, errorf("usage: %s module/path => ../local/directory", verb)
	}
	s, err := parseString(&args[0])
	if err != nil {
		return nil, errorf("invalid quoted string: %v", err)
	}

	ns, err := parseString(&args[2])
	if err != nil {
		return nil, errorf("invalid quoted string: %v", err)
	}

	if !modfile.IsDirectoryPath(ns) {
		return nil, errorf("replacement module must be directory path (rooted or starting with ./ or ../)")
	}
	if filepath.Separator == '/' && strings.Contains(ns, `\`) {
		return nil, errorf("replacement directory appears to be Windows path (on a non-windows system)")
	}

	return &modfile.Replace{
		Old:    module.Version{Path: s, Version: "v0.0.0"},
		New:    module.Version{Path: ns},
		Syntax: line,
	}, nil
}
