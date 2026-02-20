// Package validators implements the on-chain validator set management through Proof of Contribution.
// The Realm exposes a public executor for govdao proposals, that can suggest validator set changes.
// It also exposes public AddValidator/RemoveValidator methods that can be called by whitelisted realms,
// enabling use cases such as interchain security. The whitelist of allowed realms is managed through
// govdao proposals via NewPropAllowedRealmUpdateRequest.
package validators
