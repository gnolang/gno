# `keccak256` 

Keccak-256 hash (32‑byte output). Port of Go's x/crypto/sha3 (legacy Keccak variant).

## Usage
```go
data := []byte("hello world")
digest := keccak256.Hash(data) // [32]byte

h := keccak256.NewLegacyKeccak256() // streaming
h.Write([]byte("hello "))
h.Write([]byte("world"))
full := h.Sum(nil) // []byte len 32
```

## API
```go
func Hash(data []byte) [32]byte
func NewLegacyKeccak256() hash.Hash
```

`Hash` = one-shot helper. `NewLegacyKeccak256` = streaming (implements hash.Hash).

## Notes
- Fixed size: 32 bytes (256 bits).
- Uses legacy Keccak padding (differs from finalized SHA3-256 padding).