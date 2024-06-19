---
id: from-go-to-gno
---

# From Go to Gno

## Runtime comparison

TODO

## Side-by-side comparison

TODO

## Lifecycle comparison

```
      _____                     _____
     / ___/__                  / ___/__  ___
    / (_ / _ \                / (_ / _ \/ _ \
    \___/\___/                \___/_//_/\___/
+----------------+           +----------------+     |
|                |           |                |     |
|   Write app    |---------->| Write contract |     |
|                |           |                |     |
+----------------+           +----------------+     |
+----------------+           +----------------+     |  Develop as usual
|                |           |                |     |
|  Test locally  |---------->|  Test locally  |     |
|                |           |                |     |
+----------------+           +----------------+     v

+----------------+           +----------------+     |
|                |           |                |     |
|    Compile     |-----+     |                |     |
|                |     |     |                |     |
+----------------+     |     |                |     |
|                |     |     |                |     |
|  Rent hosting  |-----+     |                |     |
|                |     |     |                |     |
+----------------+     +---->|Publish on chain|     |  Deploy
|                |     |     |                |     |
| Upload binary  |-----+     |                |     |
|                |     |     |                |     |
+----------------+     |     |                |     |
|                |     |     |                |     |
|Setup a database|-----+     |                |     |
|                |           |                |     |
+----------------+           +----------------+     v

+----------------+           +----------------+     |
|   Users can    |           |                |     |
| interact with  |-----+     |                |     |
|   the server   |     |     |                |     |
+----------------+     |     |                |     |
|                |     |     |                |     |
|   Monitoring   |     |     |                |     |
|                |     |     | Users interact |     |
+----------------+     |     | with the chain |     |  Run
|    Maintain    |     |     |                |     |
|    database    |-----+---->|    Forever     |     |
|                |     |     |                |     |
+----------------+     |     |   Automatic    |     |
|                |     |     |  persistency   |     |
|  Scalability   |-----+     |                |     |
|                |     |     |                |     |
+----------------+     |     |                |     |
| Keep paying to |     |     |                |     |
|keep the service|-----+     |                |     |
|       up       |           |                |     |
+----------------+           +----------------+     v
```

## See also

- [go-gno-compatibility.md](../reference/go-gno-compatibility.md)
- ["go -> gno" presentation by Zack Scholl](https://github.com/gnolang/workshops/tree/main/presentations/2023-06-26--go-to-gno--schollz)
