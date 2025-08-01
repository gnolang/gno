# Anti-Squatting System Design for r/sys/users

## Overview

This document outlines the design for a generic anti-squatting system to prevent username squatting in the Gno ecosystem. The system implements sealed-bid auctions for high-value names and provides dispute resolution mechanisms.

## Problem Statement

The current user registration system in `r/sys/users` operates on a first-come-first-served basis with fixed pricing (1 GNOT). This allows malicious actors to:
- Squat valuable usernames (admin, gov, system, etc.)
- Register trademark names
- Hoard short or premium names for resale

## Solution Architecture

### 1. Generic Anti-Squatting Package (`p/antisquatting`)

A reusable package that can be integrated into any naming system:

```go
// Core interfaces for pluggability
type NameClassifier interface {
    IsHighValue(name string) bool
    GetMinimumBid(name string) int64
    GetAuctionDuration(name string) time.Duration
}

type DisputeResolver interface {
    CanDispute(name string, disputer std.Address) bool
    CreateDispute(name string, reason string) error
    ResolveDispute(disputeID string) (bool, error)
}

type AuctionManager interface {
    StartAuction(name string, classifier NameClassifier) error
    SubmitBid(name string, bidHash string, deposit int64) error
    RevealBid(name string, amount int64, salt string) error
    FinalizeAuction(name string) (winner std.Address, amount int64, error)
}
```

### 2. Sealed-Bid Auction Mechanism

**Commit Phase (24-168 hours configurable):**
- Users submit `hash(bid_amount + salt + address)`
- Must include minimum deposit (prevents spam)
- Multiple bids allowed (highest wins)

**Reveal Phase (24-48 hours configurable):**
- Users reveal actual bid amount and salt
- System verifies hash matches commitment
- Invalid reveals forfeit deposit

**Finalization:**
- Highest valid bid wins
- Winner pays their bid amount
- Losers get deposits refunded
- Winner gets name registered

### 3. Name Classification System

**High-Value Categories:**
- **System names**: admin, root, system, gov, dao, etc.
- **Short names**: 1-3 characters
- **Common words**: Dictionary words, brands
- **Trademark protection**: Configurable list

**Classification Logic:**
```go
func (nc *DefaultNameClassifier) IsHighValue(name string) bool {
    // Short names (1-3 chars)
    if len(name) <= 3 { return true }
    
    // System reserved words
    if isSystemReserved(name) { return true }
    
    // Common dictionary words
    if isDictionaryWord(name) { return true }
    
    // Trademark list
    if isTrademarkProtected(name) { return true }
    
    return false
}
```

### 4. Dispute Resolution Integration

**Dispute Types:**
- Trademark infringement
- Impersonation
- System name conflicts
- Bad faith registration

**Resolution Process:**
1. Dispute filed with evidence
2. DAO governance proposal created
3. Community voting period
4. Automatic execution of decision

### 5. Integration with r/sys/users

**Enhanced Registration Flow:**
```
User requests name
       ↓
Is high-value name?
   ↓         ↓
  No        Yes
   ↓         ↓
Direct    Start auction
register     ↓
   ↓      Commit phase
   ↓         ↓
   ↓      Reveal phase
   ↓         ↓
   ↓      Finalize
   ↓         ↓
   └─────────┘
       ↓
   Register winner
```

## Implementation Plan

### Phase 1: Core Package (`p/antisquatting`)
- [ ] Define interfaces and types
- [ ] Implement commit-reveal auction
- [ ] Create name classification system
- [ ] Add dispute resolution hooks

### Phase 2: Integration (`r/sys/users`)
- [ ] Enhance store with auction support
- [ ] Add anti-squatting integration layer
- [ ] Maintain backward compatibility

### Phase 3: New Implementation (`r/gnoland/users/v2`)
- [ ] Create new user realm using anti-squatting
- [ ] Implement auction-based registration
- [ ] Add dispute filing mechanisms

### Phase 4: Testing & Documentation
- [ ] Comprehensive unit tests
- [ ] Integration tests
- [ ] Local environment testing
- [ ] Documentation and examples

## Configuration Parameters

```go
type AuctionConfig struct {
    CommitDuration   time.Duration // 24h - 1 week
    RevealDuration   time.Duration // 24h - 48h
    MinimumDeposit   int64         // Spam prevention
    MinimumBid       int64         // Base price
    PremiumMultiplier int64        // For premium names
}

type DisputeConfig struct {
    DisputeFee       int64         // Cost to file dispute
    VotingPeriod     time.Duration // DAO voting time
    RequiredQuorum   int64         // Minimum votes needed
}
```

## Security Considerations

1. **Commit-Reveal Security**: Uses cryptographic hashes to prevent bid sniping
2. **Deposit Requirements**: Prevents spam and ensures serious bidders
3. **Time Locks**: Prevents manipulation during auction phases
4. **DAO Integration**: Decentralized dispute resolution
5. **Backward Compatibility**: Existing users unaffected

## Future Enhancements

1. **DNS TXT Verification**: For trademark holders
2. **Proof of Personhood**: Integration with identity systems
3. **Reputation Systems**: Weighted bidding based on reputation
4. **Automated Renewals**: Subscription-based name holding
5. **Secondary Markets**: Name transfer mechanisms

## Benefits

1. **Fair Allocation**: Market-based pricing for valuable names
2. **Spam Prevention**: Economic barriers to mass registration
3. **Dispute Resolution**: Community-driven conflict resolution
4. **Modularity**: Reusable across different naming systems
5. **Transparency**: All auctions and disputes are on-chain

This design ensures a fair, transparent, and economically efficient system for preventing username squatting while maintaining the decentralized nature of the Gno ecosystem.
