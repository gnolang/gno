# `poa` - Proof of Authority Validator Set

A Proof of Authority (PoA) validator set implementation for managing blockchain validators with simple add/remove operations and voting power controls.

## Features

- **Validator management**: Add and remove validators from the set
- **Voting power control**: Set and validate voting power for each validator
- **Address-based**: Uses standard Gno addresses for validator identification
- **Validation rules**: Enforces proper validator set constraints

## Usage

```go
import (
    "std"
    "gno.land/p/nt/poa"
    "gno.land/p/sys/validators"
)

// Create new PoA validator set
poa := poa.NewPoA()

// Add validators
validator1 := validators.Validator{
    Address:     std.Address("g1abc123..."),
    VotingPower: 10,
    PubKey:      "pubkey1",
}

err := poa.AddValidator(validator1.Address, validator1)
if err != nil {
    // Handle error (duplicate validator, invalid power, etc.)
}

// Remove validator
err = poa.RemoveValidator(validator1.Address)
if err != nil {
    // Handle error (validator not found)
}

// Check if validator exists
exists := poa.HasValidator(validator1.Address)

// Get validator info
validator, exists := poa.GetValidator(validator1.Address)

// Get all validators
validatorList := poa.GetValidators()
```

## Validation Rules

### Adding Validators
- Validator must not already exist in the set
- Voting power must be greater than 0
- Address must be valid

### Removing Validators  
- Validator must exist in the current set
- Can remove any existing validator

## API

```go
type PoA struct {
    // private fields
}

// Constructor
func NewPoA(opts ...Option) *PoA

// Validator management
func (p *PoA) AddValidator(address std.Address, validator validators.Validator) error
func (p *PoA) RemoveValidator(address std.Address) error

// Queries
func (p *PoA) HasValidator(address std.Address) bool
func (p *PoA) GetValidator(address std.Address) (validators.Validator, bool)
func (p *PoA) GetValidators() []validators.Validator
func (p *PoA) Size() int
```

## Governance Integration

```go
type GovernedPoA struct {
    poa *poa.PoA
    owner std.Address
}

func NewGovernedPoA(owner std.Address) *GovernedPoA {
    return &GovernedPoA{
        poa:   poa.NewPoA(),
        owner: owner,
    }
}

func (g *GovernedPoA) ProposeValidator(addr std.Address, votingPower int64, pubKey string) error {
    // Only owner can propose
    if std.CurrentCaller() != g.owner {
        return errors.New("unauthorized")
    }
    
    validator := validators.Validator{
        Address:     addr,
        VotingPower: votingPower,
        PubKey:      pubKey,
    }
    
    return g.poa.AddValidator(addr, validator)
}

func (g *GovernedPoA) RemoveValidator(addr std.Address) error {
    // Only owner can remove
    if std.CurrentCaller() != g.owner {
        return errors.New("unauthorized")
    }
    
    return g.poa.RemoveValidator(addr)
}
```

## Multi-Signature Validator Management

```go
type MultiSigPoA struct {
    poa *poa.PoA
    admins map[std.Address]bool
    proposals map[string]*ValidatorProposal
    threshold int
}

type ValidatorProposal struct {
    Validator validators.Validator
    Action    string // "add" or "remove"
    Votes     map[std.Address]bool
    VoteCount int
}

func (m *MultiSigPoA) ProposeAddValidator(validator validators.Validator) string {
    proposalID := generateID()
    m.proposals[proposalID] = &ValidatorProposal{
        Validator: validator,
        Action:    "add",
        Votes:     make(map[std.Address]bool),
        VoteCount: 0,
    }
    return proposalID
}

func (m *MultiSigPoA) VoteOnProposal(proposalID string, approve bool) error {
    caller := std.CurrentCaller()
    
    // Check if caller is admin
    if !m.admins[caller] {
        return errors.New("not authorized to vote")
    }
    
    proposal := m.proposals[proposalID]
    if proposal == nil {
        return errors.New("proposal not found")
    }
    
    // Record vote
    if !proposal.Votes[caller] && approve {
        proposal.Votes[caller] = true
        proposal.VoteCount++
        
        // Execute if threshold reached
        if proposal.VoteCount >= m.threshold {
            return m.executeProposal(proposal)
        }
    }
    
    return nil
}
```

## Error Handling

```go
var ErrInvalidVotingPower = errors.New("invalid voting power")
```

Common error scenarios:
- Adding validator with voting power ≤ 0
- Adding validator that already exists
- Removing validator that doesn't exist
- Invalid validator address format

## Use Cases

- **Permissioned networks**: Manage authorized validator nodes
- **Consortium blockchains**: Control who can validate transactions
- **Test networks**: Quickly add/remove validators for testing
- **Governance systems**: Implement validator voting and management
- **Network upgrades**: Coordinate validator set changes

## Integration with Gno

```go
var validatorSet *poa.PoA

func init() {
    validatorSet = poa.NewPoA()
    
    // Initialize with genesis validators
    genesisValidators := getGenesisValidators()
    for _, v := range genesisValidators {
        validatorSet.AddValidator(v.Address, v)
    }
}

func AddValidator(addr std.Address, votingPower int64, pubKey string) {
    // Add access control here
    validator := validators.Validator{
        Address:     addr,
        VotingPower: votingPower,
        PubKey:      pubKey,
    }
    
    err := validatorSet.AddValidator(addr, validator)
    if err != nil {
        panic(err)
    }
}
```

This package provides the foundation for building proof-of-authority consensus mechanisms and validator management systems in Gno applications.
