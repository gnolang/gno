package scenario

import (
	"errors"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/strings"
)

var errInvalidName = errors.New("invalid scenario name")

// Config is the base scenario configuration.
// It is not the actual scenario definition, but rather
// contains scenario metadata, such as the name and params
type Config struct {
	Name   string         `yaml:"name"`   // the unique name of the scenario
	Params map[string]any `yaml:"params"` // the input params for the scenario
}

func (c Config) Validate() error {
	// Make sure the name is non-empty and contains only ascii
	if c.Name == "" || !strings.IsASCIIText(c.Name) {
		return fmt.Errorf("%w: %q", errInvalidName, c.Name)
	}

	return nil
}
