# Roadmap

## Immediate Roadmap

 * https://github.com/gnolang/bounties/blob/main/readme.md#4-port-joeson-to-go
 * TODO

## Long Term Roadmap

### Concurrency

Initially, we don't need to implement routines because realm package functions
provide all the inter-realm functionality we need to implement rich smart
contract programming systems.  But later, for various reasons including
long-running background jobs, and parallel concurrency, Gno will implement
deterministic concurrency as well.

Determinism is supported by including a deterministic timestamp with each
channel message as well as periodic heartbeat messages even with no sends, so
that select/receive operations can behave deterministically even in the
presence of multiple channels to select from.
