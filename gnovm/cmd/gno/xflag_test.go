package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// varValue re-parses body and returns the unquoted string value of the
// package-level `var name = "..."` declaration named name, or ("", false)
// if no such declaration (with a string literal initializer) exists.
// Used to make assertions robust to exactly how the splice was applied,
// without hardcoding byte-for-byte expectations everywhere.
func varValue(t *testing.T, body, name string) (string, bool) {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.gno", body, 0)
	if err != nil {
		t.Fatalf("varValue: re-parsing patched output failed: %v\n---\n%s", err, body)
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, n := range valueSpec.Names {
				if n.Name != name || i >= len(valueSpec.Values) {
					continue
				}

				lit, ok := valueSpec.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}

				unquoted, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatalf("varValue: unquoting %s: %v", lit.Value, err)
				}

				return unquoted, true
			}
		}
	}

	return "", false
}

func TestPatchXVars(t *testing.T) {
	t.Parallel()

	t.Run("no overrides returns body unchanged", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar string = \"default\"\n"
		got, result := patchXVars("test.gno", body, nil)

		if got != body {
			t.Errorf("expected unchanged body, got:\n%s", got)
		}
		if result.Unmatched() {
			t.Errorf("expected no unmatched result for empty overrides, got: %+v", result)
		}
	})

	t.Run("simple override, explicit string type", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar string = \"default\"\n"
		got, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
		if result.Unmatched() {
			t.Errorf("expected myVar matched, got: %+v", result)
		}
	})

	t.Run("simple override, inferred type", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n"
		got, _ := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
	})

	t.Run("raw string literal initializer replaced", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = `default`\n"
		got, _ := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
	})

	t.Run("override value round-trips through quoting", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n"
		want := `has "quotes" and \backslash`
		got, _ := patchXVars("test.gno", body, map[string]string{"myVar": want})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != want {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, want)
		}
	})

	t.Run("unmatched name is reported as NotFound", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar otherVar = \"default\"\n"
		got, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "otherVar")
		if !ok || value != "default" {
			t.Errorf("otherVar = %q, ok = %v; want %q, true", value, ok, "default")
		}
		if len(result.NotFound) != 1 || result.NotFound[0] != "myVar" {
			t.Errorf("NotFound = %v, want [myVar]", result.NotFound)
		}
		if len(result.WrongKind) != 0 {
			t.Errorf("WrongKind = %v, want empty", result.WrongKind)
		}
	})

	t.Run("const is reported as WrongKind, not patched", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nconst myVar = \"default\"\n"
		got, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if !strings.Contains(got, `"default"`) || strings.Contains(got, `"override"`) {
			t.Errorf("expected const myVar to remain \"default\", got:\n%s", got)
		}
		if len(result.WrongKind) != 1 || result.WrongKind[0] != "myVar" {
			t.Errorf("WrongKind = %v, want [myVar]", result.WrongKind)
		}
	})

	t.Run("non-string var is reported as WrongKind", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = 42\n"
		_, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if len(result.WrongKind) != 1 || result.WrongKind[0] != "myVar" {
			t.Errorf("WrongKind = %v, want [myVar]", result.WrongKind)
		}
	})

	t.Run("var with no initializer is reported as WrongKind", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar string\n"
		_, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if len(result.WrongKind) != 1 || result.WrongKind[0] != "myVar" {
			t.Errorf("WrongKind = %v, want [myVar]", result.WrongKind)
		}
	})

	t.Run("grouped var block: only matching names are patched", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar (\n\ta = \"a-default\"\n\tb = \"b-default\"\n\tc = \"c-default\"\n)\n"
		got, result := patchXVars("test.gno", body, map[string]string{
			"a": "a-override",
			"c": "c-override",
		})

		for name, want := range map[string]string{
			"a": "a-override",
			"b": "b-default",
			"c": "c-override",
		} {
			value, ok := varValue(t, got, name)
			if !ok || value != want {
				t.Errorf("%s = %q, ok = %v; want %q, true", name, value, ok, want)
			}
		}
		if result.Unmatched() {
			t.Errorf("expected both a and c matched, got: %+v", result)
		}
	})

	t.Run("text resembling a var decl inside another var's raw string is left alone", func(t *testing.T) {
		t.Parallel()

		// Regression test: a raw string literal that merely *contains*
		// text resembling a top-level var declaration must never be
		// touched, since it isn't a real declaration -- only the actual
		// "var myVar = ..." decl below is.
		body := "package main\n\n" +
			"var otherVar = `\nvar myVar = \"fake, must not be touched\"\n`\n\n" +
			"var myVar = \"real default\"\n"
		got, _ := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}

		if !strings.Contains(got, `var myVar = "fake, must not be touched"`) {
			t.Errorf("expected otherVar's raw string contents to be preserved verbatim, got:\n%s", got)
		}
	})

	t.Run("local variable in a function body is left alone", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n\nfunc f() string {\n\tmyVar := \"local\"\n\treturn myVar\n}\n"
		got, _ := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("package-level myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}

		if !strings.Contains(got, `"local"`) {
			t.Errorf("expected the function-local assignment to be preserved, got:\n%s", got)
		}
	})

	t.Run("unparseable source is returned unchanged", func(t *testing.T) {
		t.Parallel()

		body := "this is not valid go source {{{"
		got, result := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if got != body {
			t.Errorf("expected unchanged body for unparseable source, got:\n%s", got)
		}
		if result.Unmatched() {
			t.Errorf("expected zero-value result for unparseable source, got: %+v", result)
		}
	})

	t.Run("line numbers after a shrunk multi-line literal are preserved", func(t *testing.T) {
		t.Parallel()

		// A 3-line raw string collapsed to a 1-line quoted string must
		// not shift the line number of code that follows it.
		body := "package main\n\n" + // lines 1-2
			"var Banner = `line one\n" + // line 3
			"line two\n" + // line 4
			"line three`\n\n" + // line 5
			"func main() {\n" + // line 6
			"\tpanic(\"boom\")\n" + // line 7
			"}\n" // line 8

		got, _ := patchXVars("test.gno", body, map[string]string{"Banner": "x"})

		fset := token.NewFileSet()
		astFile, err := parser.ParseFile(fset, "test.gno", got, 0)
		if err != nil {
			t.Fatalf("re-parsing patched output failed: %v\n---\n%s", err, got)
		}

		var panicLine int
		ast.Inspect(astFile, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if ok && ident.Name == "panic" {
				panicLine = fset.Position(call.Pos()).Line
			}
			return true
		})

		if panicLine != 7 {
			t.Errorf("panic() call moved to line %d after patching, want line 7 (unchanged):\n%s", panicLine, got)
		}
	})
}

func TestXFlag_SetAndForPackage(t *testing.T) {
	t.Parallel()

	x := newXFlag()

	if err := x.Set("main.myVar=override"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := x.Set("gno.land/r/demo/foo.Version=1.2.3"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	mainOverrides := x.forPackage("main", "main")
	if got, want := mainOverrides["myVar"], "override"; got != want {
		t.Errorf("forPackage(main, main)[myVar] = %q, want %q", got, want)
	}

	fooOverrides := x.forPackage("gno.land/r/demo/foo", "foo")
	if got, want := fooOverrides["Version"], "1.2.3"; got != want {
		t.Errorf("forPackage(gno.land/r/demo/foo, foo)[Version] = %q, want %q", got, want)
	}

	// An override for one package must not leak into another.
	if _, ok := mainOverrides["Version"]; ok {
		t.Errorf("expected Version override not to apply to the main package")
	}
	if _, ok := fooOverrides["myVar"]; ok {
		t.Errorf("expected myVar override not to apply to the foo package")
	}

	// A package whose declared name is "main" is also addressable via the
	// literal "main" pkgPath qualifier, regardless of its actual import
	// path (mirroring go build -X main.Var=...).
	otherMainOverrides := x.forPackage("some/other/import/path", "main")
	if got, want := otherMainOverrides["myVar"], "override"; got != want {
		t.Errorf("forPackage(some/other/import/path, main)[myVar] = %q, want %q", got, want)
	}
}

func TestXFlag_SetInvalid(t *testing.T) {
	t.Parallel()

	x := newXFlag()

	if err := x.Set("novalue"); err == nil {
		t.Error(`Set("novalue") expected an error, got nil`)
	}

	if err := x.Set("myVar=nopkgpath"); err == nil {
		t.Error(`Set("myVar=nopkgpath") expected an error (missing package path), got nil`)
	}

	if err := x.Set(".myVar=emptypkgpath"); err == nil {
		t.Error(`Set(".myVar=emptypkgpath") expected an error (empty package path), got nil`)
	}

	if err := x.Set("main.=noname"); err == nil {
		t.Error(`Set("main.=noname") expected an error (empty name), got nil`)
	}
}

func TestXFlag_NilString(t *testing.T) {
	t.Parallel()

	var x *xFlag
	if got := x.String(); got != "" {
		t.Errorf("(*xFlag)(nil).String() = %q, want empty string", got)
	}
}

func TestXTracker_Unmatched(t *testing.T) {
	t.Parallel()

	t.Run("all matched across multiple files", func(t *testing.T) {
		t.Parallel()

		xt := newXTracker()
		_, r1 := patchXVars("file1.gno", "package main\n\nvar a = \"1\"\n", map[string]string{"a": "x", "b": "y"})
		xt.record(r1)
		_, r2 := patchXVars("file2.gno", "package main\n\nvar b = \"2\"\n", map[string]string{"a": "x", "b": "y"})
		xt.record(r2)

		if err := xt.unmatched(map[string]string{"a": "x", "b": "y"}); err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("name never found in any file", func(t *testing.T) {
		t.Parallel()

		xt := newXTracker()
		_, r := patchXVars("file1.gno", "package main\n\nvar a = \"1\"\n", map[string]string{"a": "x", "missing": "y"})
		xt.record(r)

		err := xt.unmatched(map[string]string{"a": "x", "missing": "y"})
		if err == nil || !strings.Contains(err.Error(), "no such package-level var: missing") {
			t.Errorf("expected a 'no such package-level var: missing' error, got: %v", err)
		}
	})
}
