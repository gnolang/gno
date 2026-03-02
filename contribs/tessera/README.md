## Overview

`tessera` is an E2E gno.land (TM2) testing harness, without the in-memory node bullshit.

It is a composable and extensible testing suite capable of executing recipes against live gno.land node clusters.

## Recipes

Recipes are the backbone of the `tessera` testing harness. Recipes can mix-and-match existing scenarios to produce a
flow that needs to be verified, and working under specific conditions.

The tool adopts the following terminology:

- `Cluster` - a live (branch-version) node set (network). Every aspect of each node is configurable.
- `Scenario` - a test case (user story) that can be executed against a cluster. Think of them as legos for `Recipes`
- `Assertion` - a validation that should be performed on the cluster (after / before a scenario). They are bundled with
  `Scenarios`.
- `Invariant` - a validation that verifies an invariant condition of the cluster, independent of specific scenarios.
- `Recipe` - a declarative specification of a cluster topology and a sequence of scenarios to execute against it.
