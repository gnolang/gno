# `keccak256` - Keccak-256 hashing

Keccak-256 cryptographic hash function implementation. There is a port of Go's x/crypto/sha3 package.

## Usage

```go
// Simple hash
data := []byte("hello world")
hash := keccak256.Hash(data)
// Returns [32]byte

// Using hash.Hash interface
hasher := keccak256.NewLegacyKeccak256()
hasher.Write(data)
result := hasher.Sum(nil)
```

## API

```go
func Hash(data []byte) [32]byte
func NewLegacyKeccak256() hash.Hash
```

`Hash()` returns a 32-byte array. `NewLegacyKeccak256()` returns a `hash.Hash` interface for incremental hashing.

## Example Output

```go
data := []byte("hello")
hash := keccak256.Hash(data)
// hash = [29 87 201 85 113 215 164 4 ...]
```