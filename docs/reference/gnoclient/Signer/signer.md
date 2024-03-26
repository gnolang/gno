---
id: signer
---


# Signer

`Signer` is an interface that provides functionality for signing transactions.
The signer can be created from a local keybase, or from a bip39 mnemonic phrase.

```go
type Signer interface {
	Sign(SignCfg) (*std.Tx, error)
	Info() keys.Info
	Validate() error
}

type SignCfg struct {
    UnsignedTX     std.Tx
    SequenceNumber uint64
    AccountNumber  uint64
}
```

## `SignerFromKeybase`
`SignerFromKeybase` represents a signer created from a Gno keybase.

### [Info](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L56>)

```go
func (s SignerFromKeybase) Info() keys.Info
```

Info gets keypair information.

### [Sign](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L71>)

```go
func (s SignerFromKeybase) Sign(cfg SignCfg) (*std.Tx, error)
```

Sign implements the Signer interface for SignerFromKeybase. 

### [Validate](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L28>)

```go
func (s SignerFromKeybase) Validate() error
```

Validate checks if the signer is properly configured.

## SignerFromBip39

### [SignerFromBip39](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/signer.go#L129>)

```go
func SignerFromBip39(mnemonic string, chainID string, passphrase string, account uint32, index uint32) (Signer, error)
```

SignerFromBip39 creates an in\-memory keybase with a single default account. This can be useful in scenarios where storing private keys in the filesystem isn't feasible.

Warning: Using keys.NewKeyBaseFromDir is recommended where possible, as it is more secure.


