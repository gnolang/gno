package gen

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestGenerateDeterministic(t *testing.T) {
	a := Generate(42)
	b := Generate(42)
	if a != b {
		t.Fatal("same seed produced different programs")
	}
	if Generate(1) == Generate(2) {
		t.Fatal("different seeds produced the same program")
	}
}

func TestGenerateParses(t *testing.T) {
	for seed := uint64(1); seed <= 40; seed++ {
		src := Generate(seed)
		if _, err := parser.ParseFile(token.NewFileSet(), "main.go", src, parser.SkipObjectResolution); err != nil {
			t.Fatalf("seed %d: %v\n%s", seed, err, src)
		}
	}
}

func TestGenerateTypechecks(t *testing.T) {
	for seed := uint64(1); seed <= 80; seed++ {
		src := Generate(seed)
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, "main.go", src, 0)
		if err != nil {
			t.Fatalf("seed %d: %v\n%s", seed, err, src)
		}
		conf := types.Config{Importer: importer.Default()}
		_, err = conf.Check("main", fset, []*ast.File{file}, nil)
		if strings.Contains(src, "gnodiff:const-overflow") {
			if err == nil {
				t.Fatalf("seed %d constant-overflow mode typechecked\n%s", seed, src)
			}
			continue
		}
		if err != nil {
			t.Fatalf("seed %d: %v\n%s", seed, err, src)
		}
	}
}

func TestCatalogParses(t *testing.T) {
	seeds := RawCatalog()
	if len(seeds) < 250 {
		t.Fatalf("catalog has %d programs, want at least 250", len(seeds))
	}
	seen := map[string]bool{}
	for _, seed := range seeds {
		key := seed.Category + "/" + seed.Name
		if seen[key] {
			t.Fatalf("duplicate %s", key)
		}
		seen[key] = true
		if _, err := parser.ParseFile(token.NewFileSet(), "main.go", seed.Source, parser.SkipObjectResolution); err != nil {
			t.Fatalf("%s: %v\n%s", key, err, seed.Source)
		}
	}
}
