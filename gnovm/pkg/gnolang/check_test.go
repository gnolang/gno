package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var reErrCode = regexp.MustCompile(`^[a-z]+(-[a-z]+)*$`)

// TestErrCodesUnique: every ErrCode constant in check.go has a distinct
// kebab-case value; CodeNone is the only empty one.
func TestErrCodesUnique(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "check.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]string{}
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.CONST {
			continue
		}
		for _, s := range gd.Specs {
			vs := s.(*ast.ValueSpec)
			if id, ok := vs.Type.(*ast.Ident); !ok || id.Name != "ErrCode" {
				continue
			}
			for i, name := range vs.Names {
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok {
					t.Errorf("%s: not a literal", name.Name)
					continue
				}
				v, _ := strconv.Unquote(lit.Value)
				if name.Name == "CodeNone" {
					if v != "" {
						t.Errorf("CodeNone = %q, want empty", v)
					}
					continue
				}
				if !reErrCode.MatchString(v) {
					t.Errorf("%s = %q: not kebab-case", name.Name, v)
				}
				if prev, dup := seen[v]; dup {
					t.Errorf("%s = %q duplicates %s", name.Name, v, prev)
				}
				seen[v] = name.Name
			}
		}
	}
	if len(seen) < 30 {
		t.Fatalf("found only %d codes; is check.go parsed?", len(seen))
	}
}

// TestErrCodeTravels: a code raised by fail at a site that names no node
// reaches PreprocessError.Code() through wrapPreprocessPanic, positioned
// at the declaration it is wrapped at (the import path rule on chain).
func TestErrCodeTravels(t *testing.T) {
	m := NewMachine("main", nil)
	defer m.Release()
	fn, err := m.ParseFile("main.gno", "package main\n\nimport \"\"\n\nfunc main() {}\n")
	if err != nil {
		t.Fatal(err)
	}
	pn := NewPackageNode("main", "main", &FileSet{Files: []*FileNode{fn}})
	setNodeLines(fn)
	setNodeLocations(pn.PkgPath, fn.FileName, fn)
	d := fn.Decls[0]
	var pe *PreprocessError
	func() {
		defer func() { pe = wrapPreprocessPanic([]BlockNode{pn, fn}, d, recover()) }()
		assertValidImportPath("")
	}()
	if pe.Code() != CodeImportPath {
		t.Errorf("code: want %q, got %q", CodeImportPath, pe.Code())
	}
	if msg := pe.Unwrap().Error(); !strings.HasPrefix(msg, "main/main.gno:3:8-10: invalid import path (empty string)") {
		t.Errorf("message: %q", msg)
	}
}
