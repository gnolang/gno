# ADR: Gas Metering for Go Native Functions

## PR: #5256

## Context

Go native functions (sha256, ed25519, banker, emit, params) had flat gas costs regardless
of input size, enabling potential DoS attacks by passing large inputs that consume significant
CPU time for minimal gas cost.

## Decision

Add variable gas metering inside each native function, charging gas proportional to the
actual work done. Each function charges based on the dimension that drives its computational cost.

### Gas Model Per Function

| Function | Charging Dimension | Formula | Constant(s) |
|----------|-------------------|---------|-------------|
| `sha256.Sum256` | Per 64-byte block | `(len(data)/64 + 1) * cost` | `GasCostSha256PerBlock = 20` |
| `ed25519.Verify` | Base + per 128-byte block | `base + (len(msg)/128 + 1) * cost` | `Base = 25000, PerBlock = 60` |
| `chain.Emit` | Per byte of attr data | `totalBytes * cost` | `GasCostEmitPerByte = 1` |
| `sys/params.Set*` | Per byte of value data | `totalBytes * cost` | `GasCostParamPerByte = 1` |

### Calibration

Constants calibrated via Go benchmarks (Apple M5, `go test -bench`):
- 1 gas ≈ 1 nanosecond of CPU time (consistent with `GasFactorCPU = 1` in machine.go).
- SHA-256: ~18-20 ns per 64-byte block.
- Ed25519: ~25,500 ns fixed (EC operations) + ~63 ns per 128-byte SHA-512 block.
- Emit/Params: <1 ns/byte. Allocation gas and store gas provide primary protection.

### Design Choices

**Block-based charging for hash functions**: SHA-256 processes data in 64-byte blocks; a 1-byte
and 63-byte input both process exactly 1 block. Per-block charging matches the actual computation
granularity. Ed25519 internally uses SHA-512 (128-byte blocks) for message hashing.

**Base + per-block for Ed25519**: Verification involves two scalar multiplications on the Ed25519
curve (~25µs fixed cost) plus message hashing (O(n)). Without a base cost, small messages would
be severely undercharged relative to their actual CPU cost.

**No native gas metering for banker**: Benchmarking showed the Go-side CPU overhead of
`SendCoins` is negligible (~2-3 ns/coin for `CompactCoins`). The expensive work — KV store
reads and writes per coin denomination — is already charged by the gas-metered store
(ReadCostFlat=1000, WriteCostFlat=2000 per operation). Adding redundant native gas metering
would provide no meaningful DoS protection beyond what the store layer already guarantees.

## Alternatives Considered

- **Per-byte for all functions**: Rejected — overstates granularity for hash functions
  (sub-block inputs have identical cost).
- **Per-call flat cost**: Rejected — doesn't prevent large-input DoS for sha256/ed25519.
- **Store-layer only**: Rejected — doesn't capture CPU-intensive operations like crypto that
  bypass the KV store.

## Consequences

- Gas costs for these native functions increase from effectively 0 to proportional values.
- Ed25519 verification of small messages now correctly reflects the expensive EC operations.
- These are consensus-breaking changes that must be coordinated with chain upgrades.
