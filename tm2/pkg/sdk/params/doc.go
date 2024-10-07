// Package params provides a lightweight implementation inspired by the x/params
// module of the Cosmos SDK.
//
// It includes a keeper for managing key-value pairs with module identifiers as
// prefixes, along with a global querier for retrieving any key from any module.
//
// Changes: This version removes the concepts of subspaces and proposals,
// allowing the creation of multiple keepers identified by a provided prefix.
// Proposals may be added later when governance modules are introduced. The
// transient store and .Modified helper have also been removed but can be
// implemented later if needed. Keys are represented as strings instead of
// []byte.
package params
