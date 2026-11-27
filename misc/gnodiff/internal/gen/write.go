package gen

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"
)

// WriteCatalog formats RawCatalog and writes corpus/<category>/<name>.src.
func WriteCatalog(dir string) error {
	for _, seed := range RawCatalog() {
		formatted, err := format.Source([]byte(seed.Source))
		if err != nil {
			return fmt.Errorf("format %s/%s: %w\n%s", seed.Category, seed.Name, err, seed.Source)
		}
		path := filepath.Join(dir, seed.Category, seed.Name+".src")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, formatted, 0o644); err != nil {
			return err
		}
	}
	return nil
}
