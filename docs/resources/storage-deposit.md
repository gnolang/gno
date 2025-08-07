# Storage deposits

In Gno.land, storage is a paid resource. To persist data in a realm (such as
setting variables or storing objects), users must lock GNOT tokens as a storage
deposit. This ensures efficient, accountable use of on-chain storage.

Storage costs are settled per message, and tokens are locked or refunded
depending on the net change in data usage. Note: Gas fees are not refunded.

### What is a Storage Deposit?

A storage deposit is an amount of GNOT locked to pay for the storage space your
data occupies on-chain. The system calculates and deducts this amount after each
message (e.g., `MsgCall`, `MsgRun`, `AddPkg`).

Storing data → GNOT locked
Deleting data → GNOT refunded

### Purpose

- Paying for persistent storage: Storing objects or primitives in realms costs GNOT.
- Encouraging cleanup: Users can reclaim deposits by deleting data which is not used/needed.
- Flexibility: Realm developers can design their own cleanup or reward logic.

### Storage Settlement Flow

Below is an example of how the storage fee settlement flow works:

1. Start with a message call (e.g. `AddPkg`)
2. Specify optional `-max-deposit` to limit the GNOT that can be locked for storage.
3. The storage delta is calculated by the GnoVM (how much it grew or shrunk).
4. The system locks or refunds GNOT accordingly.

### Anyone Can Free Storage

Because deposits are tracked per realm (not per user), anyone can remove data.
This can create a “cleanup reward” scenario where the person who deletes data
gains the released deposit. It also gives the realm developer flexibility to
design and manage user storage.

### Global Storage Price Parameter

The storage price is a global parameter governed by the GovDAO.
The default value is defined in [`gno.land/pkg/sdk/vm/params.go`](https://github.com/gnolang/gno/blob/8452891dee1a92643dd0ceb4623e2c684455d3d5/gno.land/pkg/sdk/vm/params.go#L18).

```text
storagePriceDefault = "100ugnot" // cost per byte
// e.g., 1 GNOT per 10KB (≈ 1B GNOT = 10TB)
```

### Tracking Storage

You can inspect the current storage usage and deposit in a realm.
See more usage [examples](../users/interact-with-gnokey.md#vmqstorage).
