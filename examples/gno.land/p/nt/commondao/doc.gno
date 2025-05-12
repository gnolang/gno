// Package commondao provides support for custom DAO implementations.
//
// It aims to be minimal and flexible allowing the implementation of multiple DAO use cases,
// like for example standalone DAO implementation or hierarchical tree based ones.
//
// The main types that are provided are the `CommonDAO`, `Proposal`, `VotingRecord` and also
// the `MemberSet`. These and other types are the ones that allow custom DAO implementations.
//
// The package also provides a `ProposalDefinition` interface to define custom proposal types.
// An implementation of this interface is required each time a new proposal is created.
// There are also a set of other functionalities that a proposal definition can optionally have.
// These add support for proposal validation, execution or custom voting options by implementing
// `Validable`, `Executable` and `CustomizableVoteChoices` interfaces within a proposal definition.
package commondao
