# Minimal & Performant FIFO Mempool for GnoLand

This repository contains a new mempool implementation for GnoLand focused on **performance**, **simplicity**, and **extensibility**. This implementation was developed in response to [Issue #1830](https://github.com/gnolang/gno/issues/1830), which highlighted critical limitations in the legacy `CListMempool`.

---

## Motivation

The original Gno mempool (`CListMempool`) has accumulated structural complexity over time, which makes it difficult to maintain and extend. More importantly, it suffers from several protocol-level deficiencies:

* ❌ **No ordering by account sequence (nonce)** – causing invalid transaction ordering in blocks
* ❌ **No prioritization by gas price** – undermining incentive mechanisms
* ❌ **Inefficient structure for DDoS resistance or selective transaction building**

Fixing these issues in the current design would require an **extensive rewrite**. Instead, this implementation provides a **clean, minimal core** — making it ideal for experimentation and alignment with future protocol improvements.

This implementation **resolves the nonce ordering issue**, ensuring transactions from the same sender are executed in proper sequence, eliminating scenarios where dependent transactions fail due to incorrect ordering — a key part of [Issue #1830](https://github.com/gnolang/gno/issues/1830).

---

## What This Implementation Improves

* ✅ **Faster performance**: Benchmarks show **2–3× speedup** in transaction addition and update.
* ✅ **Simple FIFO logic**: Transactions are stored in the order they arrive, using an internal map and hash slice for efficient lookup and iteration.
* ✅ **Lower memory usage**: Fewer allocations and lighter data structures.
* ✅ **Correct nonce ordering**: Ensures transaction dependencies are respected by executing account transactions in proper sequence.
* ✅ **Thread-safe access** to mempool via controlled locking mechanism.
* ✅ **ABC-compliant**: Implements standard interfaces (`AddTx`, `Update`, `Flush`, `Pending`, etc.).
* ✅ **Modular & extensible**: Easy to add nonce sorting, gas prioritization, spam prevention, etc., without modifying core logic.

---

## Code Structure

```go
// my_mempool.go
Mempool struct {
  txMap      map[string]txEntry // transactions by hash
  txHashes   []string           // ordered hashes for FIFO
  proxyApp   appconn.Mempool    // ABCI app conn
  mutex      sync.RWMutex       // concurrency
  txsBytes   int64              // total tx size
}
```

* `AddTx` validates and inserts transactions
* `Update` removes committed transactions
* `Pending` returns ready txs within block limits
* `Flush`, `GetTx`, `Size`, `TxsBytes` are all lightweight helpers

---

## Benchmarks

Both implementations were tested with `go test -bench=. -benchmem` on the same machine:

* **CPU**: 13th Gen Intel Core i7-13700H
* **Go version**: 1.22
* **OS**: Linux

### Raw Benchmark Visuals

**Optimized FIFO Mempool**

![My Mempool](https://github.com/user-attachments/assets/b2f8a81d-bde6-4372-8515-928863bbf45b)


**CList Mempool**

![Clist Mempool](https://github.com/user-attachments/assets/e6af8140-8764-4f2e-9335-0e70b3de0d15)

---

### Performance Summary

| Operation        | Optimized FIFO | CList Mempool | Speedup     |
| ---------------- | -------------- | ------------- | ----------- |
| AddTx (10k txs)  | 8.58 ms        | 20.01 ms      | \~2.3×      |
| Update (10k txs) | 0.89 ms        | 2.94 ms       | \~3.3×      |
| Memory Used      | 5.8 MB         | 15.7 MB       | \~2.7× less |
| Allocations      | 50k            | 253k          | \~5× fewer  |

> While **Pending** is slightly faster in CList due to its internal queue structure, this has minimal effect on overall throughput compared to the significant gains in **AddTx** and **Update**.

✅ These results demonstrate a **significant performance advantage** in real-world scenarios.

---

## Foundation for Future Features

In agreement with mentor **Miloš Živković**, caching and recheck logic were deliberately not included in this implementation, since their future relevance is still under discussion.

A fully prioritized mempool variant — including sender nonce tracking, gas prioritization, and advanced selection logic — is available here:
[`application_mempool`](https://github.com/Milosevic02/gno/tree/feat/application_mempool/tm2/pkg/bft/app_mempool)

Due to current limitations in protocol support (e.g., lack of access to nonce or account info during mempool validation), this extended version is **not yet usable in practice**.

---

## Final Thoughts

This implementation offers more than just a working alternative — it provides a **clear, minimal, and high-performance foundation** for building the next generation of GnoLand mempool logic.

By resolving the **nonce-ordering issue** and significantly improving **transaction throughput** and **memory usage**, this design sets a new **baseline** for what a performant mempool should be.

Its **simplicity is its strength** — easy to understand, maintain, and extend.

With proven performance gains of up to **3× faster Add/Update operations** and **5× fewer allocations**, it delivers **real, measurable improvements** over the legacy `CListMempool`.

While this may not yet be a **production-ready** mempool, it can easily become one with **minimal, well-contained extensions** thanks to its clean foundation.

---

*Developed by Dragan Milošević as part of a GnoLand grant, under guidance of Miloš Živković.*
