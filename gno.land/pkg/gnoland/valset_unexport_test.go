package gnoland

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// TestValsetExportedSurface asserts the set of exported top-level
// identifiers in examples/gno.land/r/sys/validators/v3/validators.gno.
//
// This is the C1 regression test: it locks the realm's public surface
// to a known allow-list. Any change (new export, renamed export,
// re-exposed `NewValsetChangeExecutor`, etc.) fails this test and
// forces a deliberate update — better than a reject-list which would
// silently miss `NewValsetChangeExecutor2`/`MakeValsetExecutor`/etc.
//
// The unexported `newValsetChangeExecutor` is checked separately:
// it MUST be present (catches "deleted instead of renamed" regression).
func TestValsetExportedSurface(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	root := filepath.Join(wd, "..", "..", "..")
	src := filepath.Join(root, "examples", "gno.land", "r", "sys", "validators", "v3", "validators.gno")

	// .gno files parse as Go syntax for top-level decls.
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, src, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", src, err)
	}

	allowedExports := map[string]bool{
		"NewProposalRequest": true,
		"IsValidator":        true,
		"GetValidator":       true,
		"GetValidators":      true,
		"Render":             true,
	}

	requiredUnexported := map[string]bool{
		"newValsetChangeExecutor": true,
	}

	var gotExported, gotUnexported []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil { // skip methods
			continue
		}
		name := fn.Name.Name
		if ast.IsExported(name) {
			gotExported = append(gotExported, name)
		} else if requiredUnexported[name] {
			gotUnexported = append(gotUnexported, name)
		}
	}
	sort.Strings(gotExported)

	// Allow-list check.
	for _, name := range gotExported {
		if !allowedExports[name] {
			t.Errorf("unexpected exported function %q in validators.gno; "+
				"if intentional, update the allow-list", name)
		}
	}
	for name := range allowedExports {
		found := false
		for _, g := range gotExported {
			if g == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected exported function %q missing from validators.gno", name)
		}
	}

	// Required-unexported check.
	for name := range requiredUnexported {
		found := false
		for _, g := range gotUnexported {
			if g == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("required unexported function %q missing from validators.gno "+
				"(C1 regression: was it deleted instead of renamed from NewValsetChangeExecutor?)", name)
		}
	}
}
