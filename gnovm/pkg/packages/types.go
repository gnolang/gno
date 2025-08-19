package packages

import (
	"fmt"
	"go/scanner"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

type Package struct {
	Dir        string                `json:",omitempty"` // directory containing package sources
	ImportPath string                `json:",omitempty"` // import path of package in dir
	Name       string                `json:",omitempty"` // package name
	Match      []string              `json:",omitempty"` // command-line patterns matching this package
	Errors     []*Error              `json:",omitempty"` // error loading this package (not dependencies)
	Ignore     bool                  `json:",omitempty"`
	Files      FilesMap              `json:",omitempty"`
	Imports    map[FileKind][]string `json:",omitempty"` // import paths used by this package
	// XXX: Deps       []string              `json:",omitempty"` // all (recursively) imported dependencies
	// XXX: DepOnly    bool                  // package was loaded as dependency and not explicitly requested

	ImportsSpecs ImportsMap `json:"-"`
}

type Error struct {
	Pos string // "file:line:col" or "file:line" or "" or "-"
	Msg string
}

func (err Error) Error() string {
	sb := strings.Builder{}
	if err.Pos != "" {
		sb.WriteString(err.Pos)
		sb.WriteString(": ")
	}
	sb.WriteString(err.Msg)
	return sb.String()
}

func fromErr(err error, root string, prependRoot bool) []*Error {
	switch err := err.(type) {
	case scanner.ErrorList:
		res := make([]*Error, 0, len(err))
		for _, e := range err {
			pos := e.Pos.String()
			if prependRoot {
				pos = filepath.Join(root, pos)
			}
			res = append(res, &Error{Msg: e.Msg, Pos: pos})
		}
		return res
	case modfile.ErrorList:
		res := make([]*Error, 0, len(err))
		for _, e := range err {
			var pos string
			if e.Pos.LineRune > 1 {
				// Don't print LineRune if it's 1 (beginning of line).
				// It's always 1 except in scanner errors, which are rare.
				pos = fmt.Sprintf("%s:%d:%d", e.Filename, e.Pos.Line, e.Pos.LineRune)
			} else if e.Pos.Line > 0 {
				pos = fmt.Sprintf("%s:%d", e.Filename, e.Pos.Line)
			} else if e.Filename != "" {
				pos = e.Filename
			}

			var directive string
			if e.ModPath != "" {
				directive = fmt.Sprintf("%s %s: ", e.Verb, e.ModPath)
			} else if e.Verb != "" {
				directive = fmt.Sprintf("%s: ", e.Verb)
			}

			res = append(res, &Error{Msg: directive + e.Err.Error(), Pos: pos})
		}
		return res
	default:
		return []*Error{{
			Pos: root,
			Msg: fmt.Sprintf("%s (type: %T)", err.Error(), err),
		}}
	}
}

type FilesMap map[FileKind][]string

// Merge merges imports, it removes duplicates and sorts the result
func (fm FilesMap) Merge(kinds ...FileKind) []string {
	res := make([]string, 0, 16)

	for _, kind := range kinds {
		res = append(res, fm[kind]...)
	}

	sortPaths(res)
	return res
}

func sortPaths(imports []string) {
	slices.SortStableFunc(imports, func(a, b string) int {
		return strings.Compare(a, b)
	})
}
