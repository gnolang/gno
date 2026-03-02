package recipe

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/contribs/tessera/pkg/cluster"
	"github.com/gnolang/gno/contribs/tessera/pkg/scenario"
	"github.com/gnolang/gno/tm2/pkg/strings"
)

var (
	errInvalidName        = errors.New("invalid recipe name")
	errInvalidDescription = errors.New("invalid description")
	errNoScenarios        = errors.New("no scenarios in recipe")
)

// Config represents the top-level recipe configuration
type Config struct {
	Name        string            `yaml:"name"`        // the name of the recipe
	Description string            `yaml:"description"` // the description of the recipe
	Cluster     cluster.Config    `yaml:"cluster"`     // the cluster config
	Scenarios   []scenario.Config `yaml:"scenarios"`   // the recipe scenarios
}

func (c Config) Validate() error {
	// Make sure the name is non-empty and contains only ascii
	if c.Name == "" || !strings.IsASCIIText(c.Name) {
		return fmt.Errorf("%w: %q", errInvalidName, c.Name)
	}

	// Make sure the description is set
	if c.Description == "" {
		return errInvalidDescription
	}

	// Make sure there is at least 1 scenario
	if len(c.Scenarios) == 0 {
		return errNoScenarios
	}

	// Make sure the scenarios are valid
	for index, s := range c.Scenarios {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("unable to validate scenario #%d: %w", index, err)
		}
	}

	return nil
}
