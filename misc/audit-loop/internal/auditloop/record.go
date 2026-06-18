package auditloop

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Record struct {
	ID       string    `yaml:"id" json:"id"`
	Title    string    `yaml:"title" json:"title"`
	Rule     string    `yaml:"rule" json:"rule"`
	Fixtures []Fixture `yaml:"fixtures" json:"fixtures"`
}

type Fixture struct {
	Name            string `yaml:"name" json:"name"`
	Path            string `yaml:"path" json:"path"`
	WantGNOTest     string `yaml:"want_gno_test" json:"want_gno_test"`
	WantPatternHits int    `yaml:"want_pattern_hits" json:"want_pattern_hits"`
}

func LoadRecord(path string) (Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Record{}, err
	}

	var rec Record
	if err := yaml.Unmarshal(data, &rec); err != nil {
		return Record{}, err
	}
	if err := rec.validate(); err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, err)
	}

	base := filepath.Dir(path)
	for i := range rec.Fixtures {
		if !filepath.IsAbs(rec.Fixtures[i].Path) {
			rec.Fixtures[i].Path = filepath.Clean(filepath.Join(base, rec.Fixtures[i].Path))
		}
		abs, err := filepath.Abs(rec.Fixtures[i].Path)
		if err != nil {
			return Record{}, err
		}
		rec.Fixtures[i].Path = abs
	}

	return rec, nil
}

func (rec Record) validate() error {
	if rec.ID == "" {
		return fmt.Errorf("missing id")
	}
	if rec.Title == "" {
		return fmt.Errorf("missing title")
	}
	if rec.Rule == "" {
		return fmt.Errorf("missing rule")
	}
	if len(rec.Fixtures) == 0 {
		return fmt.Errorf("missing fixtures")
	}
	for i, fixture := range rec.Fixtures {
		if fixture.Name == "" {
			return fmt.Errorf("fixtures[%d]: missing name", i)
		}
		if fixture.Path == "" {
			return fmt.Errorf("fixtures[%d]: missing path", i)
		}
		switch fixture.WantGNOTest {
		case "pass", "fail":
		default:
			return fmt.Errorf("fixtures[%d]: want_gno_test must be pass or fail", i)
		}
		if fixture.WantPatternHits < 0 {
			return fmt.Errorf("fixtures[%d]: want_pattern_hits must be >= 0", i)
		}
	}
	return nil
}
