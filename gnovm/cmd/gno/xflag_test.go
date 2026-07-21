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
// Used to make assertions robust to go/printer's exact formatting choices
// (spacing, alignment, etc.), which this package doesn't want to hardcode
// expectations about.
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
		got := patchXVars("test.gno", body, nil)

		if got != body {
			t.Errorf("expected unchanged body, got:\n%s", got)
		}
	})

	t.Run("simple override, explicit string type", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar string = \"default\"\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
	})

	t.Run("simple override, inferred type", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
	})

	t.Run("raw string literal initializer replaced", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = `default`\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}
	})

	t.Run("override value round-trips through quoting", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n"
		want := `has "quotes" and \backslash`
		got := patchXVars("test.gno", body, map[string]string{"myVar": want})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != want {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, want)
		}
	})

	t.Run("no matching var name leaves declaration untouched", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar otherVar = \"default\"\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "otherVar")
		if !ok || value != "default" {
			t.Errorf("otherVar = %q, ok = %v; want %q, true", value, ok, "default")
		}
	})

	t.Run("grouped var block: only matching names are patched", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar (\n\ta = \"a-default\"\n\tb = \"b-default\"\n\tc = \"c-default\"\n)\n"
		got := patchXVars("test.gno", body, map[string]string{
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
	})

	t.Run("text resembling a var decl inside another var's raw string is left alone", func(t *testing.T) {
		t.Parallel()

		// Regression test for a review comment on the original,
		// regex-based implementation: a raw string literal that merely
		// *contains* text resembling a top-level var declaration must
		// never be touched, since it isn't a real declaration -- only
		// the actual "var myVar = ..." decl below is.
		body := "package main\n\n" +
			"var otherVar = `\nvar myVar = \"fake, must not be touched\"\n`\n\n" +
			"var myVar = \"real default\"\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}

		// otherVar's raw string content -- which merely contains text
		// that *looks* like a var myVar declaration -- must be preserved
		// byte-for-byte, proving the fake declaration inside it was
		// never treated as a real one.
		if !strings.Contains(got, `var myVar = "fake, must not be touched"`) {
			t.Errorf("expected otherVar's raw string contents to be preserved verbatim, got:\n%s", got)
		}
	})

	t.Run("local variable in a function body is left alone", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nvar myVar = \"default\"\n\nfunc f() string {\n\tmyVar := \"local\"\n\treturn myVar\n}\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		value, ok := varValue(t, got, "myVar")
		if !ok || value != "override" {
			t.Errorf("package-level myVar = %q, ok = %v; want %q, true", value, ok, "override")
		}

		if !strings.Contains(got, `"local"`) {
			t.Errorf("expected the function-local assignment to be preserved, got:\n%s", got)
		}
	})

	t.Run("const declarations are never patched", func(t *testing.T) {
		t.Parallel()

		body := "package main\n\nconst myVar = \"default\"\n"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if !strings.Contains(got, `"default"`) || strings.Contains(got, `"override"`) {
			t.Errorf("expected const myVar to remain \"default\", got:\n%s", got)
		}
	})

	t.Run("unparseable source is returned unchanged", func(t *testing.T) {
		t.Parallel()

		body := "this is not valid go source {{{"
		got := patchXVars("test.gno", body, map[string]string{"myVar": "override"})

		if got != body {
			t.Errorf("expected unchanged body for unparseable source, got:\n%s", got)
		}
	})
}

func TestXFlag_SetAndString(t *testing.T) {
	t.Parallel()

	x := newXFlag()

	if err := x.Set("myVar=override"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if got, want := x.values["myVar"], "override"; got != want {
		t.Errorf("values[myVar] = %q, want %q", got, want)
	}

	// A value containing an "=" sign should only split on the first one.
	if err := x.Set("eq=a=b"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	if got, want := x.values["eq"], "a=b"; got != want {
		t.Errorf("values[eq] = %q, want %q", got, want)
	}
}

func TestXFlag_SetInvalid(t *testing.T) {
	t.Parallel()

	x := newXFlag()

	if err := x.Set("novalue"); err == nil {
		t.Error("Set(\"novalue\") expected an error, got nil")
	}

	if err := x.Set("=novalue"); err == nil {
		t.Error("Set(\"=novalue\") expected an error, got nil")
	}
}

func TestXFlag_NilString(t *testing.T) {
	t.Parallel()

	var x *xFlag
	if got := x.String(); got != "" {
		t.Errorf("(*xFlag)(nil).String() = %q, want empty string", got)
	}
}
