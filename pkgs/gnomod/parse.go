package gnomod

import (
	"fmt"
	"reflect"
	"strconv"
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
	fss := reflect.ValueOf(fs).Interface().(*modfile.FileSyntax)
	f := &File{
		Syntax: fss,
	}
	var errs modfile.ErrorList

	for _, x := range fs.Stmt {
		switch x := x.(type) {
		case *Line:
			f.add(&errs, nil, x, x.Token[0], x.Token[1:])

		case *LineBlock:
			if len(x.Token) > 1 {
				errs = append(errs, modfile.Error{
					Filename: file,
					Pos:      modfile.Position(x.Start),
					Err:      fmt.Errorf("unknown block type: %s", strings.Join(x.Token, " ")),
				})
				continue
			}
			switch x.Token[0] {
			default:
				errs = append(errs, modfile.Error{
					Filename: file,
					Pos:      modfile.Position(x.Start),
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

func (f *File) add(errs *modfile.ErrorList, block *LineBlock, line *Line, verb string, args []string) {
	// Ignore all unknown directives
	switch verb {
	case "go", "module", "require":
		// want these even for dependency gno.mods
	default:
		return
	}

	wrapError := func(err error) {
		*errs = append(*errs, modfile.Error{
			Filename: f.Syntax.Name,
			Pos:      modfile.Position(line.Start),
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
		line := reflect.ValueOf(line).Interface().(*modfile.Line)
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
		line := reflect.ValueOf(line).Interface().(*modfile.Line)
		f.Require = append(f.Require, &modfile.Require{
			Mod:    module.Version{Path: s, Version: ""},
			Syntax: line,
		})

		// case "replace":
		// 	replace, wrappederr := parseReplace(f.Syntax.Name, line, verb, args, fix)
		// 	if wrappederr != nil {
		// 		*errs = append(*errs, *wrappederr)
		// 		return
		// 	}
		// 	f.Replace = append(f.Replace, replace)

	}
}

func parseString(s *string) (string, error) {
	t := *s
	if strings.HasPrefix(t, `"`) {
		var err error
		if t, err = strconv.Unquote(t); err != nil {
			return "", err
		}
	} else if strings.ContainsAny(t, "\"'`") {
		// Other quotes are reserved both for possible future expansion
		// and to avoid confusion. For example if someone types 'x'
		// we want that to be a syntax error and not a literal x in literal quotation marks.
		return "", fmt.Errorf("unquoted string cannot contain quote")
	}
	*s = modfile.AutoQuote(t)
	return t, nil
}
