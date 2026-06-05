# Storage Gas Calibration Benchmarks

Spec for benchmarks needed to calibrate storage gas constants in
`gnovm/pkg/gnolang/store.go` (`DefaultGasConfig`). Follows the same pattern as
the CPU gas calibration pipeline (`bench_ops_test.go` → `gen_analysis.py` →
`machine.go` constants).

See [STORAGE_CHARGING.md](STORAGE_CHARGING.md) for the current design and
[STORAGE_CHARGING_ISSUES.md](STORAGE_CHARGING_ISSUES.md) for known problems
these benchmarks should help resolve.

---

## Current State

The CPU gas pipeline is fully instrumented:

```
bench_ops_test.go → op_bench_do_dedicated.txt → gen_analysis.py → machine.go constants
alloc_bench_test.go → bench_output_do_dedicated.txt → gen_alloc_table.py → alloc.go table
```

Storage gas constants now have a partial pipeline via `gnovm/cmd/benchstore/`.

---

## Key Assumption: Amino Cost is Linear in Bytes

Benchmarks of amino serialization across diverse GnoVM value types
(see `STORAGE_CHARGING_AMINO_HEURISTIC.png`) confirm that amino encoding time
is proportional to encoded byte size. This means we do NOT need per-operation
value-size sweeps — a single per-byte constant captures amino cost. The
remaining unknowns are:

1. **Backend I/O cost** — how PebbleDB Get/Set/Delete scales with DB population
2. **Fixed per-operation overhead** — key construction, cache lookup, hash
   computation
3. **IAVL surcharge** — extra cost for escaped objects

---

## Benchmark Infrastructure

### Location

Benchmarks live in `gnovm/cmd/benchstore/` as an external test package. The
package uses the exported `Store` interface and `DefaultPebbleOptions()` from
`tm2/pkg/db/pebbledb/`.

Amino serialization benchmarks in the same package are gated behind
`//go:build genproto2` until the genproto2 amino branch lands.

### PebbleDB Configuration

Benchmarks use `pebbledb.DefaultPebbleOptions()` which matches production:
- 500 MB block cache
- Bloom filters (10 bits/key)
- 64 MB memtable
- 3 concurrent compactions

Flags allow overriding: `-cache-mb`, `-memtable-mb`, `-compactions`.
Use `-max-keys=N` to skip DB sizes above N keys.

### Running

```bash
# Full run (up to 100M keys, needs ~25GB disk):
go test -bench=. ./gnovm/cmd/benchstore/ -benchmem -timeout=30m -max-keys=100000000

# Include 1B keys (needs ~200GB disk, ~6h):
go test -bench=. ./gnovm/cmd/benchstore/ -benchmem -timeout=12h
```

---

## What Is Measured

### PebbleDB Raw I/O vs DB Size

These isolate the backing DB cost, independent of amino or GnoVM logic:

| Benchmark | What it measures |
|-----------|-----------------|
| `StorePebbleGet` | Random read latency vs DB population (1K–1B keys) |
| `StorePebbleSetOverwrite` | Random overwrite latency vs DB population |
| `StorePebbleSetInsert` | Sequential insert latency vs DB population |
| `StorePebbleDeleteAndInsert` | Delete + insert pair to keep DB size stable |

Each benchmark populates the DB, iterates the full keyspace for cache warmup,
then measures the hot loop. Setup happens once per key count (not per `b.N`
scaling).

All use 8-byte keys and 256-byte random values (incompressible to avoid
Snappy compression artifacts).

### Amino Serialization (genproto2 branch)

Behind `//go:build genproto2`:

| Benchmark | What it measures |
|-----------|-----------------|
| `AminoMarshalReflect` | Reflection-based amino encode across value types |
| `AminoMarshalBinary2` | Genproto2 direct encode |
| `AminoUnmarshalReflect` | Reflection-based amino decode |
| `AminoUnmarshalBinary2` | Genproto2 direct decode |

These produced the linear per-byte heuristic in
`STORAGE_CHARGING_AMINO_HEURISTIC.png`.

### SHA256 Hashing

In `gnovm/cmd/calibrate/sha256_bench_test.go`:

| Benchmark | Sizes |
|-----------|-------|
| `SHA256` | 32, 64, 128, 256 bytes |

Measures the hash computation cost per `SetObject` call (`HashBytes(bz)`).

---

## What Is NOT Measured (Yet)

The following are identified in STORAGE_CHARGING_ISSUES.md but not yet
benchmarked:

1. **Full GnoVM store operations at scale** — `SetObject`/`GetObject` through
   the full store path (amino + hash + backend) at varying DB sizes. The raw
   PebbleDB benchmarks give us the I/O component, and amino benchmarks give us
   the serialization component; combining them requires end-to-end store
   benchmarks.

2. **IAVL vs baseStore comparison** — escaped objects write to both stores.
   Need to measure the IAVL surcharge.

3. **Flat cost extraction** — fitting `ns = a + b × size` to determine whether
   a flat cost component should be added to `GasConfig` (P3).

4. **Read vs write ratio** — confirming the expected 2-5x write premium (P2).

---

## Reference Data

### PebbleDB I/O (Intel Xeon 8168 @ 2.70GHz, 2 cores, default PebbleOptions)

| Keys | Get (ns/op) | SetOverwrite (ns/op) | SetInsert (ns/op) | Del+Insert (ns/op) |
|------|-------------|---------------------|-------------------|-------------------|
| 1K | 403 | 839 | 813 | 1,334 |
| 10K | 490 | 1,105 | 782 | 1,459 |
| 100K | 760 | 1,018 | 671 | 1,478 |
| 1M | 1,761 | 1,031 | 671 | 1,510 |
| 10M | 4,036 | 1,530 | 654 | 1,925 |
| 100M | 27,753 | 3,569 | 787 | 4,443 |

Note: these used the old compressed values (same 256 bytes for all keys).
Re-run with unique random values gives ~10x higher Get at 100M keys (255µs)
due to realistic cache pressure on 25GB of data with 500MB cache.

### SHA256 (Apple M2)

| Size | ns/op |
|------|-------|
| 32 B | 46 |
| 64 B | 63 |
| 128 B | 89 |
| 256 B | 136 |

~46 ns base + ~0.35 ns/byte.

---

## Analysis Pipeline

### Script: `gen_storage_analysis.py`

Not yet written. Should:

1. Parse benchmark output (`ns/op` and value size / key count)
2. For amino benchmarks: confirm linear per-byte model, extract slope
3. For PebbleDB benchmarks: fit `ns = a + b × log2(keys)` to quantify
   DB-size scaling
4. Combine amino + I/O + hash costs to propose `GasConfig` constants
5. Compare proposed vs current constants

### Running on Reference Hardware

All calibration must run on the same reference machine used for CPU gas:

- **DigitalOcean Dedicated CPU Droplet**
- **Intel Xeon Platinum 8168 @ 2.70GHz, 2 cores**
- **1 gas = 1 ns**

Save reference output to `cmd/calibrate/store_bench_do_dedicated.txt`.

---

## Files

| File | Description |
|------|-------------|
| `gnovm/cmd/benchstore/store_test.go` | PebbleDB I/O benchmarks (Get/SetOverwrite/SetInsert/DeleteAndInsert × DB sizes) |
| `gnovm/cmd/benchstore/amino_test.go` | Amino marshal/unmarshal benchmarks (behind `genproto2` build tag) |
| `gnovm/cmd/benchstore/values.go` | Realistic GnoVM value constructors for amino benchmarks (behind `genproto2` build tag) |
| `gnovm/cmd/benchstore/package.go` | Package registration |
| `gnovm/cmd/calibrate/sha256_bench_test.go` | SHA256 hashing benchmarks |
| `tm2/pkg/db/pebbledb/pebbledb.go` | `DefaultPebbleOptions()` — production PebbleDB tuning |
