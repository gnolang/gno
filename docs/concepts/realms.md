---
id: realms
---

# Realms
In Gno, Realms are Gno accounts which are addressable by a 
[Gno address](../reference/standard-library/std/address.md). Realms have several 
properties:
- They can own, receive & send [Coins](./standard-library/coin.md) through the
[Banker](./standard-library/banker.md) module
- They can be part of a transaction call stack, as a caller or a callee
- They can be with or without code - EOAs, or smart contracts

Realms are represented by a `Realm` type in Gno:
```go
type Realm struct {
    addr    Address // Gno address in the bech32 format
    pkgPath string  // realm's path on-chain
}
```

## Smart Contract Realms
Often simply called "realms", Gno smart contracts contain Gno code and exist
on-chain at a specific address and package path.

## Externally Owned Accounts (EOAs)
EOAs, or simply user realms, are Gno addresses generated from a BIP39 mnemonic
phrase in a key management application, such as [`gnokey`](../gno-tooling/cli/gnokey.md).














