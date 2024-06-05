---
id: events
---

# Gno Events

## Overview

Events in Gno are a fundamental aspect of interacting with and monitoring
on-chain applications. They serve as a bridge between the on-chain environment 
and off-chain services, making it simpler for developers, analytics tools, and 
monitoring services to track and respond to activities happening in Gno.land.

Gno events are pieces of data that log specific activities or changes occurring 
within the state of an on-chain app. These activities are user-defined; they might
be token transfers, changes in ownership, updates in user profiles, and more.
Each event is recorded in the ABCI results of each block, ensuring that action 
that happened is verifiable and accessible to off-chain services. 

## Emitting Events

To emit an event, you can use the `Emit()` function from the `std` package 
provided in the Gno standard library. The `Emit()` function takes in a string 
representing the type of event, and an even number of arguments after representing
`key:value` pairs. 

Read more about events & `Emit()` in 
[Effective Gno](../08-effective-gno.md#emit-gno-events-to-make-life-off-chain-easier),
and the `Emit()` reference [here](../../07-reference/03-stdlibs/01-std/05-chain.md#emit).

## Data contained in a Gno Event

An event contained in an ABCI response of a block will include the following
data:

``` json
{
    "@type": "/tm.gnoEvent", // TM2 type
    "type": "OwnershipChange", // Type/name of event defined in Gno
    "pkg_path": "gno.land/r/demo/example", // Path of the emitter
    "func": "ChangeOwner", // Gno function that emitted the event
    "attrs": [ // Slice of key:value pairs emitted
        {
            "key": "oldOwner",
            "value": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
        },
        {
            "key": "newOwner",
            "value": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj"
        }
    ]
}
```

You can fetch the ABCI response of a specific block by using the `/block_results` 
RPC endpoint.

