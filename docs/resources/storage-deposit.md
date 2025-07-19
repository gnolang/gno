
## Storage deposits

In Gno.land, storage is a paid resource. To persist data in a realm (such as
setting variables or storing objects), users must lock GNOT tokens as a storage
deposit. This ensures efficient, accountable use of on-chain storage.

Storage costs are settled per message, and tokens are locked or refunded
depending on the net change in data usage.

### What is a Storage Deposit?

A storage deposit is an amount of GNOT locked to pay for the storage space your
data occupies on-chain. The system calculates and deducts this amount after each
message (e.g., MsgCall, MsgRun, AddPkg).

Storing data → GNOT locked
Deleting data → GNOT refunded

### Purpose

- Pay for persistent storage: Store objects like structs or strings in realms.
- Encourage cleanup: Reclaim deposits by deleting unneeded data.
- Flexibility: Realm developers can design their own cleanup or reward logic.

### Storage Settlement Flow

1 Start with a message call (e.g. AddPkg)

2 Specify optional -max-deposit to limit the GNOT to be locked for storage.

3 Storage delta is calculated (how much it grew or shrunk).

4 System locks or refunds GNOT accordingly.

### Max deposit flag

With the optional `-max-deposit flag` in gnokey, users can specify the maximum
storage deposit that may be locked when deploying a package—since the package
consumes on-chain storage—or when executing a `MsgCall` or `MsgRun`. The
transaction will fail if the chain attempts to lock more tokens than the
specified limit, protecting users from locking more tokens than they are willing
to tolerate.

### Anyone Can Free Storage

Because deposits are tracked per realm (not per user), anyone can remove data.
This can create a “cleanup reward” scenario where the person who deletes data
gains the released deposit. It also gives the realm developer flexibility to
design and manage user storage.

### Global Storage Price Parameter

The storage price is a global parameter governed by the GovDAO'
The default value is defined in gno.land/pkg/sdk/vm/params.go.
```
storagePriceDefault = "100ugnot" // cost per byte
// e.g., 1 GNOT per 10KB (≈ 1B GNOT = 10TB)
```

### Tracking Storage

We can inspect current storage usage and deposit in a realm.
It is explained [here](../user/interact-with-gnokey.md#`vm/qstorage`)

### Example

The Clear() function removes the content stored in the realm gno.land/r/foo.

```bash

gnokey maketx call \
  -pkgpath gno.land/r/foo \
  -func Clear \
  -gas-fee 1000000ugnot \
  -gas-wanted 10000000 \
  -broadcast \
  -remote https://rpc.gno.land:443 \
  -chainid staging \
  YOUR_KEY_NAME
```

You will see output similar to the following in the event:

```
GAS WANTED: 10000000
GAS USED:   291498
EVENTS: [
  {
    "type": "UnlockDeposit",
    "attrs": [
      {"key": "Deposit", "value": "2700ugnot"},
      {"key": "ReleaseStorage", "value": "27 bytes"}
    ]
  }
]
```
