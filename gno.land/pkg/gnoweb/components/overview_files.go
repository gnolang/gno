package components

import (
	"go/token"
	"slices"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/doc"
)

// FileClass holds boolean classification flags for a package file name.
// IsGno is true for test files too (they end in ".gno").
type FileClass struct {
	IsGno     bool
	IsTest    bool
	IsReadme  bool
	IsLicense bool
	IsConfig  bool
}

// ClassifyFile categorizes a file name once so callers don't re-implement the
// suffix checks. ReLicenseFileName is the shared license-name matcher.
func ClassifyFile(name string) FileClass {
	return FileClass{
		IsGno:     strings.HasSuffix(name, ".gno"),
		IsTest:    strings.HasSuffix(name, "_test.gno") || strings.HasSuffix(name, "_filetest.gno"),
		IsReadme:  name == "README.md",
		IsLicense: ReLicenseFileName.MatchString(name),
		IsConfig:  name == "gnomod.toml",
	}
}

// FileLink is one row of a package's file list (ui/file_list): the file, what
// it declares, and what that means for a reader.
type FileLink struct {
	Name string
	Link string
	FileClass

	// Summary says what the file holds in a few words; Detail is the full
	// form, for its tooltip. Both stay empty when nothing is known.
	Summary, Detail string
	// Actions counts the exported crossing functions the file declares: each
	// call to one sends a transaction.
	Actions int
	// Render is set on a realm's file that declares Render, which draws its
	// Content tab; State on one that declares package variables.
	Render, State bool
}

// maxDetailNames caps the names a file's tooltip spells out.
const maxDetailNames = 12

// fileDecls is what a package's doc says one of its files declares.
type fileDecls struct {
	top     nameList // exported top-level names
	methods nameList // exported methods of exported types, as T.M()
	actions int
	render  bool
	state   bool
}

// nameList keeps the first maxDetailNames names added, and counts them all.
type nameList struct {
	names []string
	total int
}

func (l *nameList) add(name string) {
	if l.total < maxDetailNames {
		l.names = append(l.names, name)
	}
	l.total++
}

// declsByFile groups what jdoc declares by file: types, then functions, then
// values, each in doc order. Render and state are a realm's alone: a pure
// package has neither a rendered page nor state.
func declsByFile(jdoc *doc.JSONDocumentation, realm bool) map[string]*fileDecls {
	out := map[string]*fileDecls{}
	of := func(file string) *fileDecls {
		d := out[file]
		if d == nil {
			d = &fileDecls{}
			out[file] = d
		}
		return d
	}
	for _, t := range jdoc.Types {
		if token.IsExported(t.Name) {
			of(t.File).top.add(t.Name)
		}
	}
	for _, fn := range jdoc.Funcs {
		if fn.Type != "" {
			// A method belongs to its type, and names its file only when
			// the file declares nothing at the top level.
			if token.IsExported(fn.Type) && token.IsExported(fn.Name) {
				of(fn.File).methods.add(fn.Type + "." + fn.Name + "()")
			}
			continue
		}
		d := of(fn.File)
		if ok, _ := RenderShape(fn); ok && realm {
			d.render = true
		}
		if token.IsExported(fn.Name) {
			d.top.add(fn.Name + "()")
		}
		if IsAction(fn) {
			d.actions++
		}
	}
	for _, v := range jdoc.Values {
		d := of(v.File)
		d.state = d.state || (realm && !v.Const)
		for _, vv := range v.Values {
			if token.IsExported(vv.Name) {
				d.top.add(vv.Name)
			}
		}
	}
	return out
}

// summary names the first two exported names, or for a file of methods only,
// its first two methods; detail is the full form, up to maxDetailNames names
// and a count of the rest, empty when summary says it all.
func (d *fileDecls) summary() (summary, detail string) {
	l := d.top
	if l.total == 0 {
		l = d.methods
	}
	switch {
	case l.total == 0:
		return "no exports", "Nothing exported"
	case l.total <= 2:
		return strings.Join(l.names, ", "), ""
	}
	summary = strings.Join(l.names[:2], ", ") + " +" + strconv.Itoa(l.total-2)
	detail = strings.Join(l.names, ", ")
	if l.total > len(l.names) {
		detail += " and " + strconv.Itoa(l.total-len(l.names)) + " more"
	}
	return summary, detail
}

// buildFileLinks lists the files of the package at pkgPath, each linked to its
// source view. jdoc says what each .gno file declares; nil means the doc could
// not be read, and the rows then claim nothing about declarations.
func buildFileLinks(pkgPath string, files []string, jdoc *doc.JSONDocumentation, realm bool) []FileLink {
	var decls map[string]*fileDecls
	if jdoc != nil {
		decls = declsByFile(jdoc, realm)
	}
	out := make([]FileLink, 0, len(files))
	for _, f := range files {
		l := FileLink{Name: f, Link: pkgPath + "$source&file=" + f, FileClass: ClassifyFile(f)}
		switch {
		case l.IsConfig:
			l.Summary = "Module path and Gno version"
		case l.IsReadme:
			l.Summary = "Documentation"
		case l.IsGno && !l.IsTest && jdoc != nil:
			d := decls[f]
			if d == nil {
				d = &fileDecls{}
			}
			l.Summary, l.Detail = d.summary()
			l.Actions, l.Render, l.State = d.actions, d.render, d.state
		}
		out = append(out, l)
	}
	return out
}

// IsAction reports whether fn is an action: an exported top-level crossing
// function, which a reader calls with a transaction. A file's row and a
// package's card count the same ones.
func IsAction(fn *doc.JSONFunc) bool {
	return fn.Type == "" && fn.Crossing && token.IsExported(fn.Name)
}

// RenderShape reports whether fn is a realm's Render: top-level, taking a path
// and returning the page, either as Render(string) string or, crossing, as
// Render(realm, string) string.
func RenderShape(fn *doc.JSONFunc) (ok, crossing bool) {
	if fn.Type != "" || fn.Name != "Render" || len(fn.Results) != 1 || fn.Results[0].Type != "string" {
		return false, false
	}
	switch {
	case len(fn.Params) == 1 && fn.Params[0].Type == "string":
		return true, false
	case len(fn.Params) == 2 && fn.Params[0].Type == "realm" && fn.Params[1].Type == "string":
		return true, true
	}
	return false, false
}

// DeclaresRender reports whether jdoc declares a realm's Render.
func DeclaresRender(jdoc *doc.JSONDocumentation) bool {
	if jdoc == nil {
		return false
	}
	return slices.ContainsFunc(jdoc.Funcs, func(fn *doc.JSONFunc) bool {
		ok, _ := RenderShape(fn)
		return ok
	})
}
