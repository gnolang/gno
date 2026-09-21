package integration

import (
	"fmt"
	"os"
)

// Scenario is one script and the cluster it runs against.
type Scenario struct {
	Spec ClusterSpec
	Path string
}

// ResolveScenarios reads each script's declared cluster and settles the command
// line's overrides over it.
//
// One scenario per script, so every one of them gets a chain built from genesis
// and torn down after. A chain carries deployed packages, account balances and
// its height, and none of that is reset while it runs, so two scenarios sharing
// one would make the second's result depend on the first: on what it deployed,
// on what it spent, and on the name that decided which ran first. That is a
// suite where adding a file changes the verdict of a file nobody touched.
//
// Scenarios come back in the caller's order.
func ResolveScenarios(paths []string, overrides ClusterOverrides) ([]Scenario, error) {
	scenarios := make([]Scenario, 0, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read script %s: %w", path, err)
		}
		declared, err := ParseClusterSpec(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		scenarios = append(scenarios, Scenario{Spec: overrides.Apply(declared), Path: path})
	}
	return scenarios, nil
}
