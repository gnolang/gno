# CommonDAO Package

CommonDAO is a general-purpose package that provides support to implement
custom Decentralized Autonomous Organizations (DAO) on Gno.land.

It offers a minimal and flexible framework for building DAOs, with customizable
options that adapt across multiple use cases.

## Core Types

Package contains some core types which are important in any DAO implementation,
these are **CommonDAO**, **ProposalDefinition**, **Proposal** and **Vote**.

### 1. CommonDAO Type

CommonDAO type is the main type used to define DAOs, allowing standalone DAO
creation or hierarchical tree based ones.

During creation, it accepts many optional arguments some of which are handy
depending on the DAO type. For example, standalone DAOs might use IDs, a name
and description to uniquely identify individual DAOs; Hierarchical ones might
choose to use slugs instead of IDs, or even a mix of both.

#### DAO Creation Examples

Standalone DAO:

```go
import "gno.land/p/nt/commondao"

dao := commondao.New(
    commondao.WithID(1),
    commondao.WithName("MyDAO"),
    commondao.WithDescription("An example DAO"),
    commondao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
    commondao.WithMember("g1hy6zry03hg5d8le9s2w4fxme6236hkgd928dun"),
)
```

Hierarchical DAO:

```go
import "gno.land/p/nt/commondao"

dao := commondao.New(
    commondao.WithSlug("parent"),
    commondao.WithName("ParentDAO"),
    commondao.WithMember("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
)

subDAO := commondao.New(
    commondao.WithSlug("child"),
    commondao.WithName("ChildDAO"),
    commondao.WithParent(dao),
)
```

### 2. ProposalDefinition Type

Proposal definitions are the way proposal types are implemented in `commondao`.
Definitions are required when creating a new proposal because they define the
behavior of the proposal.

Generally speaking, proposals can be divided in two types, one are the
*general* (a.k.a. *text proposals*), and the other are the *executable* ones.
The difference is that *executable* ones modify the blockchain state when they
are executed after they have been approved, while *general* ones don't, they
are usually used to signal or measure sentiment, for example regarding a
relevant issue.

Creating a new proposal type requires implementing the following interface:

```go
type ProposalDefinition interface {
    // Title returns proposal title.
    Title() string

    // Body returns proposal's body.
    // It usually contains description or values that are specific to
    // the proposal, like a description of the proposal's motivation
    // or the list of values that would be applied when the proposal
    // is approved.
    Body() string

    // VotingPeriod returns the period where votes are allowed after
    // proposal creation. It's used to calculate the voting deadline
    // from the proposal's creationd date.
    VotingPeriod() time.Duration

    // Tally counts the number of votes and verifies if proposal passes.
    // It receives a voting context containing a readonly record with the votes
    // that has been submitted for the proposal and also the list of DAO members.
    Tally(VotingContext) (passes bool, _ error)
}
```

This minimal interface is the one required for *general proposal types*. Here
the most important method is the `Tally()` one. It's used to check whether a
proposal passes or not.

Within `Tally()` votes can be counted using different rules depending on the
proposal type, some proposal types might decide if there is consensus by using
super majority while others might decide using plurality for example, or even
just counting that a minimum number of certain positive votes have been
submitted to approve a proposal.

CommonDAO provides a couple of helpers for this, to cover some cases:
- `SelectChoiceByAbsoluteMajority()`
- `SelectChoiceBySuperMajority()` (using a 2/3s threshold)
- `SelectChoiceByPlurality()`

#### 2.1. Executable Proposals

Proposal definitions have optional features that could be implemented to extend
the proposal type behaviour. One of those is required to enable execution
support.

A proposal can be executable implementing the **Executable** interface as part
of the new proposal definition:

```go
type Executable interface {
    // Executor returns a function to execute the proposal.
    Executor() func(realm) error
}
```

The crossing function returned by the `Executor()` method is where the realm
changes are made once the proposal is executed.

Other features can be enabled by implementing the **Validable** interface and
the **CustomizableVoteChoices** one, as a way to separate pre-execution
validation and to support proposal voting choices different than the default
ones (YES, NO and ABSTAIN).

### 3. Proposal Type

Proposals are key for governance, they are the main mechanic that allows DAO
members to engage on governance.

They are usually not created directly but though **CommonDAO** instances, by
calling the `CommonDAO.Propose()` or `CommonDAO.MustPropose()` methods. Though,
alternatively, proposals could be added to CommonDAO's active proposals storage
using `CommonDAO.ActiveProposals().Add()`.

```go
import (
    "gno.land/p/nt/commondao"
    "gno.land/r/example/mydao"
)

dao := commondao.New()
creator := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
propDef := mydao.NewGeneralProposalDefinition("Title", "Description")
proposal := dao.MustPropose(creator, propDef)
```

#### 3.1. Voting on Proposals

The preferred way to submit a vote, once a proposal is created, is by calling
the `CommonDAO.Vote()` method because it performs sanity checks before a vote
is considered valid; Alternatively votes can be directly added without sanity
checks to the proposal's voting record by calling
`Proposal.VotingRecord().AddVote()`.

#### 3.2. Voting Record

Each proposal keeps track of their submitted votes within an internal voting
record. CommonDAO package defines it as a **VotingRecord** type.

The voting record of a proposal can be getted by calling its
`Proposal.VotingRecord()` method.

Right now proposals have a single voting record but the plan is to support
multiple voting records per proposal as an optional feature, which could be
used in cases where a proposal must track votes in multiple independent
records, for example in cases where a proposal could be promoted to a different
DAO with a different set of members.

#### 4. Vote Type

Vote type defines the structure to store information for individual proposal
votes. Apart from the normally mandatory `Address` and voting `Choice` fields,
there are two optional fields that can be useful in different use cases; These
fields are `Reason` which can store a string with the reason for the vote, and
`Context` which can be used to store generic values related to the vote, for
example vote weight information.

It's *very important* to be careful when using the `Context` field, in case
references/pointers are assigned to it because they could potentially be
accessed anywhere, which could lead to unwanted indirect modifications.

Vote type is defined as:

```go
type Vote struct {
    // Address is the address of the user that this vote belons to.
    Address address

    // Choice contains the voted choice.
    Choice VoteChoice

    // Reason contains an optional reason for the vote.
    Reason string

    // Context can store any custom voting values related to the vote.
    Context any
}
```

## Secondary Types

There are other types which can be handy for some implementations which might
require to store DAO members or proposals in a custom location, or that might
need member grouping support.

### 1. MemberStorage and ProposalStorage Types

These two types allows storing and iterating DAO members and proposals. They
support DAO implementations that might require storing either members or
proposals in an external realm other than the DAO realm.

CommonDAO package provides implementations that use AVL trees under the hood
for storage and lookup.

Custom implementations are supported though the **MemberStorage** and
**ProposalStorage** interfaces:

```go
type MemberStorage interface {
	// Size returns the number of members in the storage.
	Size() int

	// Has checks if a member exists in the storage.
	Has(address) bool

	// Add adds a member to the storage.
	Add(address) bool

	// Remove removes a member from the storage.
	Remove(address) bool

	// Grouping returns member groups when supported.
	Grouping() MemberGrouping

	// IterateByOffset iterates members starting at the given offset.
	IterateByOffset(offset, count int, fn func(address) bool)
}

type ProposalStorage interface {
	// Has checks if a proposal exists.
	Has(id uint64) bool

	// Get returns a proposal or nil when proposal doesn't exist.
	Get(id uint64) *Proposal

	// Add adds a proposal to the storage.
	Add(*Proposal)

	// Remove removes a proposal from the storage.
	Remove(id uint64)

	// Size returns the number of proposals that the storage contains.
	Size() int

	// Iterate iterates proposals.
	Iterate(offset, count int, reverse bool, fn func(*Proposal) bool) bool
}
```

### 2. MemberGrouping and MemberGroup Types

Members grouping is an optional feature that provides support for DAO members
grouping.

Grouping can be useful for DAOs that require grouping users by roles or tiers
for example.

The **MemberGrouping** type is a collection of member groups, while the
**MemberGroup** is a group of members with metadata.

#### Grouping by Role Example

```go
import "gno.land/p/nt/commondao"

storage := commondao.NewMemberStorageWithGrouping()

// Add a member that doesn't belong to any group
storage.Add("g1...a")

// Create a member group for owners
owners, err := storage.Grouping().Add("owners")
if err != nil {
  panic(err)
}

// Add a member to the owners group
owners.Members().Add("g1...b")

// Add voting power to owners group metadata
owners.SetMeta(3)

// Create a member group for moderators
moderators, err := storage.Grouping().Add("moderators")
if err != nil {
  panic(err)
}

// Add voting power to moderators group metadata
moderators.SetMeta(1)

// Add members to the moderators group
moderators.Members().Add("g1...c")
moderators.Members().Add("g1...d")
```
