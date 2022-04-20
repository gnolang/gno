# Roadmap

This is a work in progress. For the latest discussion, see the main README.md
for link to Discord app.

## Immediate Roadmap

 * Go precompile, gnodev
 * /r/users -> /r/names; /r/boards updates. 
 * Tweak/enforce gas limitations

## Mid Term Roadmap

 * float-as-struct support
 * goroutines and concurrency
 * https://github.com/gnolang/bounties/blob/main/readme.md#4-port-joeson-to-go

## Long Term Roadmap

 * privacy-preserving voting
 * sybil resistant proof-of-human
 * open hardware 
 * logos browser

# Notes

## Concurrency

Initially, we don't need to implement routines because realm package functions
provide all the inter-realm functionality we need to implement rich smart
contract programming systems.  But later, for various reasons including
long-running background jobs, and parallel concurrency, Gno will implement
deterministic concurrency as well.

Determinism is supported by including a deterministic timestamp with each
channel message as well as periodic heartbeat messages even with no sends, so
that select/receive operations can behave deterministically even in the
presence of multiple channels to select from.
