# Analysis: Fee model trade-offs in gno.land

This document covers gas fee and storage deposit trade-offs for contract data structures on gno.land, based on observations during Akkadia development.

## Fee Components

| Cost Component | Description | When Incurred | Paid By | Recoverable |
|---|---|---|---|---|
| Gas Fee | Computation and bandwidth cost for transaction execution | Every transaction | Transaction sender | No |
| Storage Deposit | Deposit for on-chain storage allocation | When new storage is allocated | Transaction sender | Yes, when storage is freed |

## Key Trade-offs

- **Gas fees** scale with computation complexity. Heavy iteration over large state structures increases gas cost per transaction.
- **Storage deposits** scale with state size. Contracts that allocate many entries (maps, arrays) incur higher upfront storage costs but can reclaim deposits when state is freed.
- For large-scale contracts (e.g. on-chain worlds), the dominant cost driver is storage allocation, not computation. Developers should minimize state entries and prefer compact representations.

## Implications for Contract Design

- Prefer fewer, larger state entries over many small entries (reduces storage deposit count).
- Batch operations where possible to amortize gas overhead across multiple state changes.
- Consider state cleanup patterns (freeing unused entries) to reclaim storage deposits.
- Gas fees are non-recoverable — optimize computation paths for frequently called functions.
