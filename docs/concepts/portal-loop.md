---
id: portal-loop
---

# Portal Loop

The Portal Loop is an always-up-to-date staging testnet that allows for using
the latest version of Gno, Gno.land, and TM2. By utilizing the power of Docker
& the [tx-archive](https://github.com/gnolang/tx-archive) tool, the Portal Loop can run the latest code from the 
master branch on the [Gno monorepo](https://github.com/gnolang/gno), while preserving most/all the previous the transaction data. 

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

The Portal Loop deployment can be found on [portal.gno.land:port](https://portal.gno.land). 
