---
id: portal-loop
---

# Portal Loop

Portal Loop is an always-up-to-date staging testnet that allows for using
the latest version of Gno, Gno.land, and TM2. By utilizing the power of Docker
& the [tx-archive](https://github.com/gnolang/tx-archive) tool, the Portal Loop can run the latest code from the 
master branch on the [Gno monorepo](https://github.com/gnolang/gno), 
while preserving most/all of the previous transaction data. 

The Portal Loop allows for quick iteration on the latest version of Gno - without
having to make a hard/soft fork. 

Below is a diagram demonstrating how the Portal Loop works:
```
                    +----------------------------------+
                    |       Portal Loop running        |  < ----+ 
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |   Detect changes in 'master'     |        |
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    | Archive transaction data & state |        |    
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |    Load changes from 'master'    |        |
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |      Replay transaction data     |  ------+  
                    +----------------------------------+
```

Specifically, Portal Loop behaves like a normal network until a change is detected
in the `master` branch in the Gno monorepo. At this point, Portal Loop archives 
on-chain data using the [tx-archive](https://github.com/gnolang/tx-archive)
tool, saving all transactions that happened on it thus far.

It then pulls the latest changes from the `master` branch, and inserts all 
previously archived transactions into the genesis of the newly deployed chain.
After genesis has been replayed, the chain continues working as normal.

## Using the Portal Loop

The Portal Loop deployment can be found at [gno.land](https://gno.land), while
the exposed RPC endpoints can be found on `https://rpc.gno.land:443`. The RPC endpoint
list can be found in the [reference section](../reference/rpc-endpoints.md).

### A warning note

While allowing for quick iteration on the most up-to-date software, the Portal Loop
has some drawbacks:
- If a breaking change happens on `master`, transactions that used the previous version of
Gno will fail to be replayed, meaning **data will be lost**. 
- Since transactions are archived and replayed during genesis, 
block height & timestamp cannot be relied upon.
