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
// re-introducing a top-level executor like the legacy
// `NewValsetChangeExecutor`) fails this test and forces a deliberate
// update — better than a reject-list which would silently miss
// `NewValsetChangeExecutor2`/`MakeValsetExecutor`/etc.
//
// The operator-keyed builder + executor live in proposal.gno
// (NewValidatorProposalRequest + newValoperChangeExecutor); the
// signing-keyed legacy NewProposalRequest was removed (every valid
// signing-keyed input is also a valid operator-keyed input under
// always-on valoper enforcement). Tests against proposal.gno cover
// the executor's lifecycle.
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
		"IsValidator":   true,
		"GetValidator":  true,
		"GetValidators": true,
		"Render":        true,
	}

	var gotExported []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil { // skip methods
			continue
		}
		name := fn.Name.Name
		if ast.IsExported(name) {
			gotExported = append(gotExported, name)
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
}
