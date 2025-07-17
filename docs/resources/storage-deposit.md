
## Overview of storage deposit

In Gno.land, storage is a paid resource. To persist data in a realm (such as
setting variables or storing objects), users must lock GNOT tokens as a storage
deposit. This ensures efficient, accountable use of on-chain storage.

Storage costs are settled per message, and tokens are locked or refunded
depending on the net change in data usage.

## How it works

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

### Tracking Storage

Use this command to inspect current storage usage and deposit in a realm:

```bash
gnokey query vm/qstorage --data gno.land/r/foo
```

Sample Output:

```
storage: 5025, deposit: 502500

```

storage: total bytes used

deposit: total GNOT locked for that storage

### Anyone Can Free Storage

Because deposits are tracked per realm (not per user), anyone can remove data.
This can create a “cleanup reward” scenario where the person who deletes data
gains the released deposit. It also gives the realm developer flexibility to
design and manage user storage.

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
