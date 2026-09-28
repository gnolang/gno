# misc/e2e — End-to-End Test Suite

Docker-based E2E test suite for gnoland. Tests run against a single local
validator node and are executed automatically by `make test`.

## Running

```sh
cd misc/e2e
make test      # build images, start node, run all tests
make clean     # tear down containers and volumes
make logs      # stream container logs
```

## Structure

```
misc/e2e/
├── run_tests.sh          # main entrypoint called by docker-compose
├── docker-compose.yml    # spins up gnoland + gnokey-test containers
├── audit/
│   ├── common.sh         # shared config: RPC, chainid, key setup
│   └── audit_*.sh        # one script per gnovm fix (see below)
└── e2e/
    └── e2e_*.sh          # end-to-end transaction and consensus tests
```

## Audit scripts (`audit/`)

Each script targets a specific gnovm bugfix and verifies it is present in
the binary. Scripts exit 0 on ✅ PATCHED and exit 1 on ❌ VULNERABLE.

| Script | Fix | What it verifies |
| --- | --- | --- |
| `audit_runtime_pkg.sh` | `afd7e4808` | `runtime` import rejected in production VM |
| `audit_chan_type.sh` | `4bcd9828e` | `chan` type rejected at preprocess, not at runtime |
| `audit_security.sh` | `6a6fc4c71` + `3be0408f0` | uint64 overflow caught at compile time; infinite recursion stopped by gas limit |
| `audit_gas_alloc.sh` | `5d5f9213f` | large allocations consume gas proportionally (per-byte model) |
| `audit_byteslice.sh` | `a3a356e71` | byte-slice index mutations persist across transactions |
| `audit_array_alias.sh` | `c64feef1d` | array copy produces independent memory (no pointer aliasing) |
| `audit_var_init_order.sh` | `50ee56e64` | package-level vars initialized in dependency order |
| `audit_cross_realm_recover.sh` | `f87249327` | full state rollback when a realm panics and recover() is called |

## E2E scripts (`e2e/`)

| Script | What it verifies |
| --- | --- |
| `e2e_nonce_replay.sh` | Replaying a tx with an already-consumed sequence number is rejected |
| `e2e_counter.sh` | Deploy a realm, increment state, verify committed value |
| `e2e_mempool_stress.sh` | 10 sequential txs accepted without error; final state matches expected count |

## Shared config (`audit/common.sh`)

All scripts source `audit/common.sh` which sets:

| Variable | Default | Description |
| --- | --- | --- |
| `RPC` | `http://gnoland:26657` | Node RPC endpoint |
| `CHAINID` | `test` | Chain ID |
| `KEY` | `test1` | Gnokey account name |
| `PASSWORD` | `test1234` | Key password |
| `GNOKEY_HOME` | `/tmp/gnokey` | Gnokey home directory |
| `KEY_ADDR` | `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5` | Deterministic address for `test1` |

Override any variable via environment:
```sh
RPC=http://localhost:26657 CHAINID=test-13 ./audit/audit_security.sh
```
