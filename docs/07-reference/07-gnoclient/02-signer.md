---
id: signer
---

# Signer

`Signer` is an interface that provides functionality for signing transactions.
The signer can be created from a local keybase, or from a bip39 mnemonic phrase.

Useful types and functions when using the `Signer` can be found below.

## type [Signer](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L13-L17>)

`Signer` provides an interface for signing transactions.

```go
type Signer interface {
    Sign(SignCfg) (*std.Tx, error) // Signs a transaction and returns a signed tx ready for broadcasting.
    Info() keys.Info               // Returns key information, including the address.
    Validate() error               // Checks whether the signer is properly configured.
}
```

## type [SignCfg](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L65-L69>)

`SignCfg` provides the signing configuration, containing the unsigned transaction 
data, account number, and account sequence.

```go
type SignCfg struct {
    UnsignedTX     std.Tx
    SequenceNumber uint64
    AccountNumber  uint64
}
```

## type [SignerFromKeybase](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L20-L25>)

`SignerFromKeybase` represents a signer created from a Keybase.

```go
type SignerFromKeybase struct {
    Keybase  keys.Keybase // Stores keys in memory or on disk
    Account  string       // Account name or bech32 format
    Password string       // Password for encryption
    ChainID  string       // Chain ID for transaction signing
}
```

### func \(SignerFromKeybase\) [Info](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L56>)

```go
func (s SignerFromKeybase) Info() keys.Info
```

`Info` gets keypair information.

### func \(SignerFromKeybase\) [Sign](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L72>)

```go
func (s SignerFromKeybase) Sign(cfg SignCfg) (*std.Tx, error)
```

`Sign` implements the Signer interface for SignerFromKeybase.

### func \(SignerFromKeybase\) [Validate](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L28>)

```go
func (s SignerFromKeybase) Validate() error
```

`Validate` checks if the signer is properly configured.

## func [SignerFromBip39](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L130>)

```go
func SignerFromBip39(mnemonic string, chainID string, passphrase string, account uint32, index uint32) (Signer, error)
```

`SignerFromBip39` creates a `Signer` from an in-memory keybase with a single default 
account, derived from the given mnemonic.
This can be useful in scenarios where storing private keys in the filesystem
isn't feasible, or for generating a signer for testing.

> Using `keys.NewKeyBaseFromDir()` to get a keypair from local storage is 
recommended where possible, as it is more secure.