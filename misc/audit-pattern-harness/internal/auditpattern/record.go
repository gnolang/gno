package auditpattern

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
	Repair   Repair    `yaml:"repair" json:"repair"`
	Fixtures []Fixture `yaml:"fixtures" json:"fixtures"`
}

type Repair struct {
	FromFixture         string   `yaml:"from_fixture" json:"from_fixture"`
	ToFixture           string   `yaml:"to_fixture" json:"to_fixture"`
	Goal                string   `yaml:"goal" json:"goal"`
	AllowRemovedExports []string `yaml:"allow_removed_exports,omitempty" json:"allow_removed_exports,omitempty"`
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

	if err := rec.ValidatePaths(); err != nil {
		return Record{}, fmt.Errorf("%s: %w", path, err)
	}

	return rec, nil
}

// ValidatePaths checks that every fixture path exists and is a directory.
func (rec Record) ValidatePaths() error {
	for i, fixture := range rec.Fixtures {
		info, err := os.Stat(fixture.Path)
		if err != nil {
			return fmt.Errorf("fixtures[%d] %s: %w", i, fixture.Name, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("fixtures[%d] %s: path is not a directory", i, fixture.Name)
		}
	}
	return nil
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
	if rec.Repair.FromFixture == "" {
		return fmt.Errorf("repair: missing from_fixture")
	}
	if rec.Repair.ToFixture == "" {
		return fmt.Errorf("repair: missing to_fixture")
	}
	if rec.Repair.Goal == "" {
		return fmt.Errorf("repair: missing goal")
	}
	if len(rec.Fixtures) == 0 {
		return fmt.Errorf("missing fixtures")
	}
	fixtures := map[string]bool{}
	for i, fixture := range rec.Fixtures {
		if fixture.Name == "" {
			return fmt.Errorf("fixtures[%d]: missing name", i)
		}
		if fixtures[fixture.Name] {
			return fmt.Errorf("fixtures[%d]: duplicate name %q", i, fixture.Name)
		}
		fixtures[fixture.Name] = true
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
	if !fixtures[rec.Repair.FromFixture] {
		return fmt.Errorf("repair: from_fixture %q does not match a fixture", rec.Repair.FromFixture)
	}
	if !fixtures[rec.Repair.ToFixture] {
		return fmt.Errorf("repair: to_fixture %q does not match a fixture", rec.Repair.ToFixture)
	}
	if rec.Repair.FromFixture == rec.Repair.ToFixture {
		return fmt.Errorf("repair: from_fixture and to_fixture must differ")
	}
	return nil
}
