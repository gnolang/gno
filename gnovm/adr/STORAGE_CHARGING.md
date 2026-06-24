# Storage Charging

> **HISTORICAL.** This document describes the storage-gas model that existed
> *before* the gas refactor. The per-operation constants discussed here
> (`GasGetObject`, `GasSetType`, `GasGetPackageRealm`, ...) were replaced with
> two universal amino compute constants (`GasAminoEncode` / `GasAminoDecode`)
> plus a depth-aware I/O model charged at the `cache.Store` boundary. See
> `gno.land/adr/gas_refactor.md` for the current design. This file is kept for
> historical context.

Spec for the current storage cost model. Two independent mechanisms charge for
storage: **storage gas** (burned per-transaction) and **storage deposits**
(locked/refunded over the lifetime of stored data).

---

## 1. Storage Gas (GnoVM layer)

### Overview

When the GnoVM store reads, writes, or deletes objects, types, realm metadata,
or mem-packages, it charges gas proportional to the serialized byte size of the
data. This gas is consumed from the transaction's `GasMeter` and is
**non-refundable** — it covers the computational and I/O cost of the storage
operation itself.

### Gas constants

Defined in `gnovm/pkg/gnolang/store.go` (`DefaultGasConfig()`):

| Operation | Constant | Cost | Unit |
|-----------|----------|------|------|
| GetObject | `GasGetObject` | 16 | per byte |
| SetObject | `GasSetObject` | 16 | per byte |
| GetType | `GasGetType` | 5 | per byte |
| SetType | `GasSetType` | 52 | per byte |
| GetPackageRealm | `GasGetPackageRealm` | 524 | per byte |
| SetPackageRealm | `GasSetPackageRealm` | 524 | per byte |
| AddMemPackage | `GasAddMemPackage` | 8 | per byte |
| GetMemPackage | `GasGetMemPackage` | 8 | per byte |
| DeleteObject | `GasDeleteObject` | 3715 | flat |

All per-byte costs are `constant × len(amino_serialized_bytes)`.
`DeleteObject` is a flat cost (no per-byte component).

### Where gas is consumed

Each store method serializes (or deserializes) the data and charges gas on the
serialized size before writing to (or after reading from) the backend:

- **`loadObjectSafe`** (`store.go:442`): `GasGetObject × len(bz)` after reading
  from `baseStore.Get`.
- **`SetObject`** (`store.go:594`): `GasSetObject × len(bz)` after
  `amino.MustMarshalAny`. The value written to `baseStore` is `hash || bz`
  (HashSize + amino bytes). For escaped objects, a separate `oid → hash` entry
  is written to `iavlStore`.
- **`DelObject`** (`store.go:698`): flat `GasDeleteObject`. Deletes from
  `baseStore` (and `iavlStore` if escaped).
- **`GetTypeSafe`** (`store.go:731`): `GasGetType × len(bz)`.
- **`SetType`** (`store.go:781`): `GasSetType × len(bz)`.
- **`GetPackageRealm`** (`store.go:372`): `GasGetPackageRealm × len(bz)`.
- **`SetPackageRealm`** (`store.go:398`): `GasSetPackageRealm × len(bz)`.
- **`AddMemPackage`** (`store.go:911`): `GasAddMemPackage × len(bz)`.
- **`GetMemPackage`** (`store.go:946`): `GasGetMemPackage × len(bz)`.

Gas is consumed via `ds.consumeGas()` which calls `ds.gasMeter.ConsumeGas()`
if a gas meter is set (tests may run without one).

### Relationship to SDK base store gas

The SDK provides a separate gas-metering store wrapper (`tm2/pkg/store/gas/store.go`)
with its own cost schedule (`ReadCostFlat: 1000`, `WriteCostPerByte: 30`, etc.).
Modules like `auth` use this via `ctx.GasStore(key)`.

**The VM keeper does not use this wrapper.** It calls `ctx.Store(key)` (raw,
unwrapped) when creating the GnoVM transaction store
(`gno.land/pkg/sdk/vm/keeper.go:334`). The raw stores are passed directly to
`gnoStore.BeginTransaction()`. Therefore the SDK-level store gas costs
**do not apply** to GnoVM storage operations — only the GnoVM `GasConfig`
constants above are charged.

### Two backend stores

The GnoVM store uses two underlying stores, both accessed without SDK gas wrapping:

- **`baseStore`**: stores objects (as `hash || amino_bytes`), types, block nodes,
  realm metadata, and mem-packages.
- **`iavlStore`**: stores `oid → hash` mappings for escaped objects (objects
  referenced across realm boundaries), enabling IAVL Merkle proofs.

---

## 2. Storage Deposits (gno.land layer)

### Overview

Storage deposits are an economic mechanism separate from gas. When a transaction
increases a realm's on-disk storage, the caller must lock a deposit proportional
to the byte increase. When storage is later freed, the deposit is refunded. This
creates an ongoing cost for persistent state, unlike gas which is a one-time
burn.

### Parameters

Defined in `gno.land/pkg/sdk/vm/params.go`, governable via param proposals:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `StoragePrice` | `100ugnot` | Cost per byte of storage |
| `DefaultDeposit` | `600000000ugnot` | Fallback deposit cap if `msg.MaxDeposit` is 0 |
| `StorageFeeCollector` | derived from `"storage_fee_collector"` | Address that receives withheld refunds |

At the default price: 1 GNOT = 10 KB of storage.

### How storage differences are tracked

1. During realm finalization (`realm.go:387–390`), each realm accumulates a
   net byte delta (`sumDiff`) from `SetObject` and `DelObject` calls:
   - `SetObject` returns `len(hash+bz) - LastObjectSize` (the difference from
     the previous serialized size, or the full size for new objects).
   - `DelObject` returns `LastObjectSize` (the full size being removed).
   - `rlm.sumDiff += store.SetObject(oo)` for creates/updates.
   - `rlm.sumDiff -= store.DelObject(do)` for deletes.

2. The accumulated `sumDiff` is written to `store.RealmStorageDiffs()`, a
   `map[string]int64` mapping realm path to net byte change, then reset to 0.

### Deposit processing

After each VM message (`MsgAddPackage`, `MsgCall`, `MsgRun`), the keeper calls
`processStorageDeposit()` (`keeper.go:1221–1311`):

1. Read `realmDiffs` from the store.
2. Determine available deposit: `msg.MaxDeposit` if provided, else
   `params.DefaultDeposit`.
3. Sort realm paths for deterministic processing.
4. For each realm with a non-zero diff:

**If diff > 0 (storage increased):**
```
requiredDeposit = diff × StoragePrice.Amount
```
- If `depositAmt < requiredDeposit`, the transaction fails with an error.
- Otherwise, `lockStorageDeposit()` transfers `requiredDeposit` from the caller
  to a per-realm deposit address (`DeriveStorageDepositCryptoAddr(rlm.Path)`).
- `rlm.Deposit` and `rlm.Storage` are incremented.
- A `StorageDepositEvent` is emitted.

**If diff < 0 (storage decreased):**
```
depositUnlocked = |diff| × StoragePrice.Amount
```
- `refundStorageDeposit()` transfers `depositUnlocked` from the per-realm
  deposit address back to the caller.
- **Exception**: if `ugnot` is a restricted denom (token lock period), the
  refund goes to `StorageFeeCollector` instead of the caller.
- `rlm.Deposit` and `rlm.Storage` are decremented.
- A `StorageUnlockEvent` is emitted.

### Realm state

Each `Realm` (`realm.go:117–124`) tracks:

| Field | Type | Description |
|-------|------|-------------|
| `Deposit` | `uint64` | Total deposit locked for this realm (in ugnot) |
| `Storage` | `uint64` | Total storage used by this realm (in bytes) |
| `sumDiff` | `int64` | Transient accumulator for current tx's byte delta |

Queryable via `vm/qstorage` → `QueryStorage(ctx, pkgPath)`.

---

## 3. How the Two Mechanisms Interact

A single `SetObject` call participates in both systems:

1. **Gas burned**: `GasSetObject × len(amino_bytes)` is consumed from the
   transaction's gas meter. This is non-refundable and covers the cost of
   serialization and I/O.

2. **Deposit tracked**: the return value `len(hash+bz) - LastObjectSize` feeds
   into `realm.sumDiff`, which later drives `diff × StoragePrice` deposit
   locking in `processStorageDeposit`.

The two are independently parameterized:
- **Gas** covers the one-time execution cost of performing the storage operation.
- **Deposits** cover the ongoing cost of occupying persistent state, and are
  refunded when the state is freed.

A transaction that reads without modifying state pays only gas (no deposit
change). A transaction that deletes objects pays gas for the delete operation
but receives a deposit refund.

---

## 4. Cost Asymmetries

### Reads vs writes

`GasGetObject` and `GasSetObject` are both 16/byte — reads and writes are
charged equally at the GnoVM layer. This differs from the SDK gas config
(unused by VM) where writes (30/byte + 2000 flat) are 10x more expensive
than reads (3/byte + 1000 flat).

### Types vs objects

`GasSetType` (52/byte) is 3.25x more expensive than `GasSetObject` (16/byte).
Types are set once during package loading and read many times, so the higher
write cost amortizes.

### Realm metadata

`GasGetPackageRealm` and `GasSetPackageRealm` are both 524/byte — by far the
most expensive per-byte cost. Realm metadata includes deposit/storage accounting
and is read/written on every realm-touching transaction.

### Deletes

`GasDeleteObject` is a flat 3715 gas regardless of object size. There is no
per-byte component. The deposit system handles the economic incentive to free
storage (via refunds).

---

## 5. What Is Not Covered by Storage Gas

- **CPU gas for serialization/deserialization**: amino marshal/unmarshal CPU
  time is not separately metered — it is implicitly bundled into the per-byte
  storage gas constants.
- **IAVL tree rebalancing**: when escaped objects are written to `iavlStore`,
  the IAVL tree may rebalance. This cost is not separately charged.
- **Cache operations**: reads from `cacheObjects`/`cacheTypes` (in-memory
  maps) are not gas-metered. Only backend store hits incur storage gas.
