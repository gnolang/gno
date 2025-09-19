package scenario

import (
	"context"
	"errors"
	"fmt"

	"github.com/gnolang/gno/contribs/tessera/pkg/cluster"
)

var ErrNotFound = errors.New("scenario not found")

// Scenario is a test case (user story) that can be executed against a cluster.
// Essentially, they are lego blocks for Recipes
type Scenario interface {
	// Description returns the explanation of the scenario (human-readable)
	Description() string

	// Execute executes the scenario on the given (live) cluster
	Execute(context.Context, *cluster.Cluster) error

	// Verify is a validation that should be performed on the cluster after execution
	Verify(context.Context, *cluster.Cluster) error
}

// Load loads the given scenario with the params from the global registry
func Load(name string, params any) (Scenario, error) {
	// Fetch the scenario if it exists
	savedScenario, found := globalRegistry.Get(name)
	if !found {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}

	// Set up the params for the scenario
	// TODO

	return savedScenario, nil
}
