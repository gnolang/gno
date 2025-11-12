# daocond: Stateless Condition Engine for DAO Governance

`daocond` is a Gnolang package that provides a stateless condition engine for evaluating DAO proposal execution. It serves as the decision-making core of the daokit framework, determining whether proposals should be executed based on configurable governance rules. 

## Core Interfaces

### Condition Interface

```go
type Condition interface {
    // Eval - checks if the condition is satisfied based on current votes
    Eval(ballot Ballot) bool
    
    // Signal - returns a value from 0.0 to 1.0 indicating progress toward satisfaction
    Signal(ballot Ballot) float64
    
    Render() string // returns a representation of the condition
    RenderWithVotes(ballot Ballot) string // returns a representation with vote context
}
```

### Ballot Interface

```go
type Ballot interface {
    Vote(voter string, vote Vote) // allows a user to vote on a proposal
    Get(voter string) Vote // returns the vote of a specific user
    
    Total() int // returns the total number of votes cast
    
    Iterate(fn func(voter string, vote Vote) bool) // iterates over all votes
}
```

### Vote Types

```go
type Vote int

const (
    VoteAbstain Vote = iota  // Neutral vote
    VoteNo                   // Against the proposal
    VoteYes                  // In favor of the proposal
)
```

## Built-in Conditions

`daocond` provides three core condition types for common governance scenarios:

```go
// MembersThreshold - Requires a fraction of all DAO members to approve
func MembersThreshold(threshold float64, isMemberFn func(string) bool, membersCountFn func() uint64) Condition

// RoleThreshold - Requires a percentage of role holders to approve  
func RoleThreshold(threshold float64, role string, hasRoleFn func(string, string) bool, roleCountFn func(string) uint32) Condition

// RoleCount - Requires a minimum number of role holders to approve
func RoleCount(count uint64, role string, hasRoleFn func(string, string) bool) Condition
```

**Usage Examples**:
```go
// Require 60% of all members
memberMajority := daocond.MembersThreshold(0.6, store.IsMember, store.MembersCount)

// Require 50% of contributor  
adminApproval := daocond.RoleThreshold(0.5, "contributor", store.HasRole, store.RoleCount)

// Require at least 2 core-contributor
treasurerApproval := daocond.RoleCount(2, "core-contributor", store.HasRole)
```

## Logical Composition

Combine conditions using logical operators to create complex governance rules:

```go
// And - All conditions must be satisfied
func And(conditions ...Condition) Condition {...}

// Or - At least one condition must be satisfied  
func Or(conditions ...Condition) Condition {...}
```

**Examples**:
```go
// Require BOTH admin majority AND treasurer approval
strictGovernance := daocond.And(
    daocond.RoleThreshold(0.5, "contributor", store.HasRole, store.RoleCount),
    daocond.RoleCount(1, "treasurer", store.HasRole),
)

// Require EITHER treasurer majority OR unanimous core-contributor approval
flexibleGovernance := daocond.Or(
    daocond.RoleThreshold(0.5, "treasurer", store.HasRole, store.RoleCount),
    daocond.RoleThreshold(1.0, "core-contributor", store.HasRole, store.RoleCount),
)
```

## Creating Custom Conditions

Implement the `Condition` interface to create custom governance rules:

```go
type customCondition struct {
    // Your custom fields
}

func (c *customCondition) Eval(ballot daocond.Ballot) bool {
    // Implement your evaluation logic
    return true
}

func (c *customCondition) Signal(ballot daocond.Ballot) float64 {
    // Return progress from 0.0 to 1.0
    return 0.5
}

func (c *customCondition) Render() string {
    return "Custom condition description"
}

func (c *customCondition) RenderWithVotes(ballot daocond.Ballot) string {
    return "Custom condition with current vote status"
}
```

## Usage Examples

### Basic Usage

```go
import "gno.land/p/samcrew/daocond"

// Create a simple majority condition
condition := daocond.MembersThreshold(0.5, store.IsMember, store.MembersCount)

// Evaluate the condition against a ballot
if condition.Eval(ballot) {
    // Proposal meets the condition requirements
    executeProposal()
}

// Check progress toward satisfaction
progress := condition.Signal(ballot) // Returns 0.0 to 1.0
```

### Complex Governance Rules

```go
// Multi-tier approval system
governance := daocond.And(
    // Require 30% of all members
    daocond.MembersThreshold(0.3, store.IsMember, store.MembersCount),
    
    // AND at least 2 core-contributor approvals
    daocond.RoleCount(2, "core-contributor", store.HasRole),
    
    // AND either CTO approval OR finance team majority
    daocond.Or(
        daocond.RoleCount(1, "CTO", store.HasRole),
        daocond.RoleThreshold(0.5, "finance", store.HasRole, store.RoleCount),
    ),
)
```

### Integration with daokit

`daocond` is designed to work seamlessly with:
- **[daokit](/p/samcrew/daokit/)**: Core DAO framework
- **[basedao](/p/samcrew/basedao/)**: Member and role management
- Custom DAO implementations

```go
import (
    "gno.land/p/samcrew/daocond"
    "gno.land/p/samcrew/daokit"
)

// Define conditions for different types of proposals
treasuryCondition := daocond.And(
    daocond.RoleCount(1, "treasurer", store.HasRole),
    daocond.MembersThreshold(0.6, store.IsMember, store.MembersCount),
)

// Use in resource registration
resource := daokit.Resource{
    Handler:     treasuryHandler,
    Condition:   treasuryCondition,
    DisplayName: "Treasury Management",
    Description: "Proposals for treasury operations",
}
```

For complete examples and interactive demos, see the [/r/samcrew/daodemo/custom_condition](/r/samcrew/daodemo/custom_condition) realms.

---

*Part of the daokit framework for building decentralized autonomous organizations in gnolang.*