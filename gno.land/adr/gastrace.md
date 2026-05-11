# ADR: Build-Tag Store Gas Tracing

## Problem

Gas tracing for store I/O requires adding temporary `fmt.Fprintf` calls,
rebuilding, running, then removing them. This is error-prone and we've
needed it multiple times for optimization work.

## Scope

Traces **store-level gas only**: cache store I/O (Get/Set/Delete), GnoVM
amino encode/decode for objects/types/realms/mempackages, and direct IAVL
operations for escaped objects and mempackages.

Does NOT trace VM compute gas (CPU cycles, memory, parsing), ante handler
gas (txSize, sig verify), or block gas meter. For a typical realm call,
store gas accounts for ~40-70% of total gas.

## Solution

A small `tm2/pkg/store/trace` package with a build-tag const and trace
function. Both `tm2/pkg/store/cache` and `gnovm/pkg/gnolang` import it.
Zero overhead in production builds.

### Package: `tm2/pkg/store/trace`

`trace.go` (production — no build tag overhead):

```go
//go:build !gastrace

package trace

const StoreGasEnabled = false

func Store(op string, gas int64, key []byte, valLen int, info string) {}
func Flush()                                                           {}
```

`trace_on.go` (tracing build):

```go
//go:build gastrace

package trace

import (
    "bufio"
    "encoding/hex"
    "fmt"
    "io"
    "os"
)

const StoreGasEnabled = true

var out *bufio.Writer // nil when writing to stderr (unbuffered)
var outFile *os.File  // always set

func init() {
    path := os.Getenv("GAS_TRACE")
    if path == "" || path == "1" || path == "true" {
        outFile = os.Stderr
        // No bufio for stderr — crash-safe, traces visible immediately.
    } else {
        f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
        if err != nil {
            panic("GAS_TRACE: " + err.Error())
        }
        outFile = f
        out = bufio.NewWriter(f)
    }
}

func Store(op string, gas int64, key []byte, valLen int, info string) {
    keyHex := hex.EncodeToString(key)
    if len(keyHex) > 160 {
        keyHex = keyHex[:160] + "..."
    }
    keyStr := make([]byte, len(key))
    for i, b := range key {
        if b >= 0x20 && b < 0x7f {
            keyStr[i] = b
        } else {
            keyStr[i] = '.'
        }
    }
    if len(keyStr) > 80 {
        keyStr = append(keyStr[:80], '.', '.', '.')
    }
    var w io.Writer = outFile
    if out != nil {
        w = out
    }
    fmt.Fprintf(w,
        "GAS_STORE op=%-14s gas=%-10d vlen=%-6d info=%-16s key_hex=%s key_str=%s\n",
        op, gas, valLen, info, keyHex, keyStr)
}

// Flush writes buffered trace data. No-op for stderr (unbuffered).
// Must be called before os.Exit — defers do not run on os.Exit.
func Flush() {
    if out != nil {
        out.Flush()
    }
}
```

### Import and usage rules

No circular imports — `trace` is a leaf package importing only stdlib:
- `tm2/pkg/store/cache` imports `tm2/pkg/store/trace` (sibling)
- `gnovm/pkg/gnolang` imports `tm2/pkg/store/trace` (leaf dependency)

**Guard requirement:** Every call site MUST be guarded:

```go
if trace.StoreGasEnabled {
    trace.Store("GET", gas, key, len(value), "depth=false")
}
```

Without the guard, argument expressions are still evaluated even though
the no-op body is inlined. With the guard, the compiler eliminates the
entire block including argument evaluation. (Same pattern as the existing
`gnovm/pkg/benchops` package.)

**Flush:** The buffer is flushed automatically after every `TxEnd`.
No manual flush or shutdown hook is needed. At most one tx worth of
data is lost if the process crashes mid-transaction. Stderr output
is unbuffered and crash-safe regardless.

### Trace points

**tm2/pkg/store/cache/store.go** — cache I/O gas:

| Op | Where | Info | Notes |
|----|-------|------|-------|
| `GET` | Get cache miss | `depth=true/false` | Gas = total charged (flat + per-byte sum). `depth=true` when `store.hasEstimator` (IAVL-backed, DepthReadFlat). `depth=false` for flat stores (ReadFlat). |
| `SET` | Set | `depth=true/false` | Gas = total charged (read + write + per-byte sum). Same depth mapping. |
| `REFUND` | Set/Delete dedup | `dedup` | Gas value is positive, represents gas returned to meter. `vlen=0`. |
| `DELETE` | Delete | `depth=true/false` | Same depth mapping. |

**gnovm/pkg/gnolang/store.go** — amino encode/decode gas:

| Op | Where | Info | Notes |
|----|-------|------|-------|
| `DECODE_OBJ` | loadObjectSafe | `cached=true/false` | `cached=true`: loaded from stdlibKeyBytes (no I/O gas). `cached=false`: loaded from baseStore (I/O gas charged separately via GET). |
| `ENCODE_OBJ` | SetObject | `none` | |
| `DECODE_TYPE` | GetTypeSafe | `none` | |
| `ENCODE_TYPE` | SetType | `none` | |
| `DECODE_REALM` | GetPackageRealm | `none` | |
| `ENCODE_REALM` | SetPackageRealm | `none` | |
| `DECODE_MEMPKG` | getMemPackage | `none` | |
| `ENCODE_MEMPKG` | AddMemPackage | `none` | |

**gnovm/pkg/gnolang/store.go** — direct IAVL ops (bypass cache store):

| Op | Where | Info | Notes |
|----|-------|------|-------|
| `IAVL_SET_ESCAPED` | SetObject | `none` | Escaped object hash write to IAVL. |
| `IAVL_DEL_ESCAPED` | DelObject | `none` | Escaped object hash delete from IAVL. |
| `IAVL_SET_MEMPKG` | AddMemPackage | `none` | MemPackage path write to IAVL. |
| `IAVL_GET_MEMPKG` | getMemPackage | `none` | MemPackage path read from IAVL. |

The bulk copy in `CopyFromCachedStore` (test utility) is excluded.

### Usage

```bash
# Trace to stderr (unbuffered, crash-safe):
go test -tags gastrace ./gno.land/pkg/integration/ -run TestTestdata/save_struct -v

# Trace to file (buffered, Flush before exit):
GAS_TRACE=/tmp/trace.txt go test -tags gastrace ./gno.land/pkg/integration/ -run TestTestdata/save_struct

# Normal build — zero overhead:
go test ./gno.land/pkg/integration/ -run TestTestdata/save_struct
```

### Output format

Fixed 6 fields per line, `key=value` pairs, both hex and ASCII keys:

```
GAS_STORE op=DECODE_OBJ      gas=1233       vlen=431    info=cached=true     key_hex=6f69643a... key_str=oid:013551...:1
GAS_STORE op=GET              gas=59000      vlen=431    info=depth=false     key_hex=6f69643a... key_str=oid:013551...:1
GAS_STORE op=REFUND           gas=225658     vlen=0      info=dedup           key_hex=2f612f...   key_str=/a/...
GAS_STORE op=IAVL_SET_ESCAPED gas=223880     vlen=20     info=none            key_hex=30333162... key_str=031b10...:4
```

## Files to create/change

1. `tm2/pkg/store/trace/trace.go` — `StoreGasEnabled = false` + no-op stubs
2. `tm2/pkg/store/trace/trace_on.go` — `StoreGasEnabled = true` + `Store()` + `Flush()` + `init()`
3. `tm2/pkg/store/cache/store.go` — GET/SET/REFUND/DELETE trace points
4. `gnovm/pkg/gnolang/store.go` — DECODE/ENCODE + IAVL trace points

## Limitations

- **Store I/O and amino only.** Does not trace VM compute, ante handler,
  or block gas meter. Traced gas sums to ~40-70% of total GAS USED.
- **No running total.** Cross-reference with GAS USED output for totals.
- **Single-goroutine only.** The `bufio.Writer` is not mutex-protected.
  Do not use with parallel test execution and file output.
- **Crash safety.** File output is buffered and flushed after each tx.
  At most one tx of trace data is lost on crash. Stderr is unbuffered.
- **Key format.** Assumes store keys are printable ASCII (true for all
  current formats: oid:, tid:, pkg:, /a/, /pv/). Non-printable bytes
  replaced with `.` in key_str.

## Non-goals

- No runtime enable/disable (build tag only)
- No structured logging (plain text, grep/awk friendly)
- No performance profiling (just store gas accounting)
