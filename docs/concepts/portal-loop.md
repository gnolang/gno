---
id: portal-loop
---

# Portal Loop

The Portal Loop is an always-up-to-date staging testnet that allows for using
the latest version of Gno, Gno.land, and TM2. By utilizing the power of Docker
& the [tx-archive](https://github.com/gnolang/tx-archive) tool, the Portal Loop 
can stay up to date with the master branch on the [Gno monorepo](https://github.com/gnolang/gno),
while preserving all the previous the transaction data. 

Below is a diagram displaying how the Portal Loop works:
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

