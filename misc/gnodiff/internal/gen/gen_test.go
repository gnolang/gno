package gen

import (
	"go/parser"
	"go/token"
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
