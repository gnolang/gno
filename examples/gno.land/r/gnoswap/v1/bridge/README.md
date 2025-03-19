# Bridge

## Overview

Importing a package-specific path (e.g., `v1/somecontract`) embeds the path (hardcodes it) within the contract implementation. If you upgrade or redeploy a new contract version (e.g., `v2/somecontract`), existing contracts remain locked to the old path (i.e., they continue calling the `v1` version).

To solve this, `bridge` provides a callback-based function call. Instead of directly importing the contract, dependent contracts call a `bridge` contract, which holds function pointers to the currently registered contract version. When we upgrade the contract, we simply update the callbacks in the `bridge`, ensuring that all dependent contracts automatically call the new version without requiring a direct import.

## How does it work?

Below is a simple representation of how the calls flow:



```plain
        (position)
            |
            | calls `MintAndDistributeGnsCallback()`
            v
       +------------+
       |  r/bridge  |
       +-----+------+
             |
    [fn ptrs]|
             |
             v
      v1/emission    v2/emission   ... (any future version)
       (register)     (register)            (register)
```

## Disclaimer

The current implementation is experimental and serves to verify the callback pattern for upgradability. The structure and function registration method may change in the future.
