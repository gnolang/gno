> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `poa` - Proof of Authority validator set

Stateful Proof of Authority validator set with simple add/remove constraints. This is a low-level building block intended to be embedded by chain-level governance code (e.g. a GovDAO bridge to `gno.land/p/sys/validators`), not a typical realm utility.

Constraints:
- **Add**: validator must not be in the set already and voting power must be `> 0`.
- **Remove**: validator must be in the set.

## Usage

```go
import (
    "gno.land/p/nt/poa/v0"
    "gno.land/p/sys/validators"
)

// Start with a pre-seeded validator set.
set := poa.NewPoA(poa.WithInitialSet([]validators.Validator{
    {Address: "g1...", PubKey: "gpub1...", VotingPower: 10},
}))

// Add a validator.
v, err := set.AddValidator("g1xyz...", "gpub1xyz...", 5)
if err != nil {
    panic(err)
}

// Inspect membership.
if set.IsValidator("g1xyz...") {
    // ...
}

// List the full current set.
all := set.GetValidators()

// Remove a validator.
removed, err := set.RemoveValidator(v.Address)
```

## API

```go
type PoA struct { /* ... */ }

// Construct an empty set; options seed initial validators.
func NewPoA(opts ...Option) *PoA

// WithInitialSet seeds the validator set at construction time.
func WithInitialSet(vs []validators.Validator) Option

func (p *PoA) AddValidator(addr address, pubKey string, power uint64) (validators.Validator, error)
func (p *PoA) RemoveValidator(addr address) (validators.Validator, error)
func (p *PoA) IsValidator(addr address) bool
func (p *PoA) GetValidator(addr address) (validators.Validator, error)
func (p *PoA) GetValidators() []validators.Validator
```

Validators are stored and returned as `validators.Validator` from `gno.land/p/sys/validators`.

## Errors

- `ErrInvalidVotingPower` — `AddValidator` called with `power == 0`.
- `validators.ErrValidatorExists` — adding an address already in the set.
- `validators.ErrValidatorMissing` — removing or fetching an address that is not in the set.

## Notes

- Public keys are stored as-is — there is no on-chain verification yet (`TODO` in source).
- The package is intentionally narrow: it only manages the in-memory set. Consensus-layer wiring (proposing/applying changes to the actual validator set) is the caller's responsibility.
