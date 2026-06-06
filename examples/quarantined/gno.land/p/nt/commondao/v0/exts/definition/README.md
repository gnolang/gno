# CommonDAO Package Definition Extension

Definition package is an extension of `gno.land/p/nt/commondao/v0` that provides
an alternative approach to define custom proposal types.

## Definition

The `Definition` type is an implementation that allows creating custom proposal
definitions using callback functions and definition options.

CommonDAO package supports different proposal types through the
`ProposalDefinition` interface, so new proposal types require the definition
of a custom type that implements the interface. The `Definition` type is a
callback based alternative to the type based approach.

By default, new definitions have a voting period of 7 days, allowing _YES_,
_NO_ and _ABSTAIN_ votes, tallying those votes using an absolute majority of
more than 50% of member votes, considering that a proposal passes when the
majority of the votes are _YES_.

New definition can be created using any of the following functions:

```go
// New creates a new custom proposal definition or returns an error
func New(title, body string, options ...Option) (Definition, error)

// MustNew creates a new custom proposal definition or panics on error
func MustNew(title, body string, options ...Option) Definition
```

Default definition behavior can be configured by setting custom options that
configures the following proposal options:

- Custom vote choices
- Voting period
- Tally behavior
- Pre-execution and render validation
- Execution behavior

Example usage:

```go
import (
  "chain/runtime"
  "errors"
  "time"

  "gno.land/p/nt/commondao/v0"
  "gno.land/p/nt/commondao/v0/exts/definition"
)

var dao = commondao.New()

// CreateMemberProposal creates a new example proposal to add a DAO member.
func CreateMemberProposal(member address) uint64 {
  if !member.IsValid() {
    panic("invalid member address")
  }

  // Define a function to validate that member doesn't exist within the DAO
  validate := func() error {
    if dao.Members().Has(member) {
      return errors.New("member already exists within the DAO")
    }
    return nil
  }

  // Define a custom tally function that approves proposals without votes
  tally := func(commondao.VotingContext) (bool, error) {
    return true, nil
  }

  // Define an executor to add the new member to the DAO
  executor := func(realm) error {
    dao.Members().Add(member)
    return nil
  }

  // Create a custom proposal definition for an example proposal type
  def := definition.MustNew(
    "Example Proposal",
    "This is a simple proposal example",
    definition.WithVotingPeriod(time.Hour * 24 * 2), // 2 days
    definition.WithTally(tally),
    definition.WithValidation(validate),
    definition.WithExecutor(executor),
  )

  // Create a new proposal
  p := dao.MustPropose(runtime.PreviousRealm().Address(), def)
  return p.ID()
}
```
