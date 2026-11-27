package gen

import (
	"go/format"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogCheckedIn(t *testing.T) {
	dir := filepath.Join("..", "..", "corpus")
	for _, seed := range RawCatalog() {
		formatted, err := format.Source([]byte(seed.Source))
		if err != nil {
			t.Fatalf("format %s/%s: %v", seed.Category, seed.Name, err)
		}
		path := filepath.Join(dir, seed.Category, seed.Name+".src")
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing %s (run go run . -write-corpus corpus): %v", path, err)
		}
		if string(body) != string(formatted) {
			t.Fatalf("%s does not match the catalog; run go run . -write-corpus corpus", path)
		}
	}
}
