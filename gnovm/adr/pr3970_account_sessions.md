# ADR: Account Sessions Support (PR #3970)

## Status

Draft - Exploration 1

## Context

Gno accounts currently use a single key pair, creating several issues:

**Security issues**: Single key compromise means total account loss. No way to
delegate limited permissions. Key rotation requires moving all assets.

**Usability issues**: Can't automate specific tasks without exposing master key.
No temporary access for services. Multi-device usage requires sharing private
keys.

This PR explores adding multiple keys per account with different permissions.

## Design

`BaseAccount` now contains `BaseAccountKey` for both the master key and
sessions. `GnoAccount` extends this with a `Sessions []GnoSession` field. The
root key remains the master key with full control. Sessions are limited keys
with specific permissions controlled by flags:

- `flagSessionUnlimitedTransferCapacity` - unlimited transfer amount
- `flagSessionCanManageSessions` - can add/remove other sessions  
- `flagSessionCanManagePackages` - can deploy packages
- `flagSessionValidationOnly` - validator-only operations

Sessions can have transfer capacity limits that decrease with use (transfers
are allowed if capacity is non-zero or unlimited flag is set), realm access
whitelists using glob patterns, and expiration times. The garbage collection
(`gc()`) is implemented in `gno.land` layer, not `tm2`, so appchains can
customize expiration logic for their specific needs.

**Key-Account Relationship**: Since preventing key reuse across accounts is
inefficient, we allow one key to control multiple accounts and one account to
have multiple keys. This works because transactions always specify the account
address, so we know which account a signature belongs to.

**Sequence Numbers**: Each pubkey has its own independent sequence number to
prevent replay attacks. The account also maintains a global sequence sum
(`SequenceSum`) that tracks total operations across all keys. Sessions can
start at any sequence number (including 0), allowing wallets to implement
custom patterns for replay protection. The flexible sequence system lets
wallets see total account activity and choose appropriate starting sequences.

The implementation maintains backward compatibility by falling back to message
signers when signatures lack public keys (genesis transactions).

## Usage

- **Validator keys**: Validation-only sessions for node operators
- **Hot wallets**: Limited daily-use sessions while master key stays offline
- **Service accounts**: Realm-restricted sessions for automated systems
- **Temporary access**: Time-limited sessions for specific operations

## TODO

- [x] high-level changes (structs, interfaces)
- [x] vmkeeper changes
- [x] tm2 vs gno.land vs ...
- [ ] within contracts
- [x] storage efficiency
- [x] lookup efficiency
- [ ] reusing burned
- [x] sequence numbers
- [x] account vs session
- [x] account session vs subaccount vs session account + compatibility with this in the future
- [x] one account, several keys
- [x] one key, several accounts
- [ ] key rotation
- [x] new error codes
- [ ] inter-wallet
- [x] self-expiring
- [ ] expected usage
- [ ] patterns
- [ ] offline
- [ ] masterkey usage warning
- [x] constraints / filters
- [ ] validator limited keys