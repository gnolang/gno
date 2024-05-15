---
id: tm2-js-provider
---

# Overview

A `Provider` is an interface that abstracts the interaction with the Tendermint2 chain, making it easier for users to
communicate with it. Rather than requiring users to understand which endpoints are exposed, what their return types are,
and how they are parsed, the `Provider` abstraction handles all of this behind the scenes. It exposes useful API methods
that users can use and expects concrete types in return.

Currently, the `tm2-js-client` package provides support for two Provider implementations:

- [JSON-RPC Provider](json-rpc-provider.md): executes each call as a separate HTTP RPC call.
- [WS Provider](ws-provider.md): executes each call through an active WebSocket connection, which requires closing when
  not needed anymore.

## Account Methods

### getBalance

Fetches the denomination balance of the account

#### Parameters

* `address` **string** the bech32 address of the account
* `denomination` **string** the balance denomination (optional, default `ugnot`)
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<number\>**

#### Usage

```ts
await provider.getBalance('g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq', 'atom');
// 100
```

### getAccountSequence

Fetches the account sequence

#### Parameters

* `address` **string** the bech32 address of the account
* `height` **number** the height for querying.
  If omitted, the latest height is used. (optional, default `0`)

Returns **Promise<number\>**

#### Usage

```ts
await provider.getAccountSequence('g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq');
// 42
```

### getAccountNumber

Fetches the account number. Errors out if the account
is not initialized

#### Parameters

* `address` **string** the bech32 address of the account
* `height` **number** the height for querying.
  If omitted, the latest height is used (optional, default `0`)

Returns **Promise<number\>**

#### Usage

```ts
await provider.getAccountNumber('g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq');
// 100
```

## Block methods

### getBlock

Fetches the block at the specific height, if any

#### Parameters

* `height` **number** the height for querying

Returns **Promise<BlockInfo\>**

#### Usage

```ts
await provider.getBlock(1);
/*
{
  block_meta: {
    block_id: {
      hash: "TxHKEGxFm/4+D7gxOJdVUaR+xTDZzlPrCVXuVm7SqHw=",
      parts: {
        total: "1",
        hash: "+dqI9oyngnnlKyno7y+RxCLEPA9FxWA/MmXyJ4uoJAY="
      }
    },
    header: {
      version: "v1.0.0-rc.0",
      chain_id: "dev",
      height: "1",
      time: "2023-05-01T10:32:20.807541Z",
      num_txs: "0",
      total_txs: "0",
      app_version: "",
      last_block_id: {
        hash: null,
        parts: {
          total: "0",
          hash: null
        }
      },
      last_commit_hash: null,
      data_hash: null,
      validators_hash: "FnuBaDvDLg4FotGRcZAFvhLkEjkb+kNLaAZrAVhL5Aw=",
      next_validators_hash: "FnuBaDvDLg4FotGRcZAFvhLkEjkb+kNLaAZrAVhL5Aw=",
      consensus_hash: "uKhnXFmGUkxgQSJf17ogbYLNXDo3UEPwQvzddo4Vkuw=",
      app_hash: null,
      last_results_hash: null,
      proposer_address: "g1vsqzyy9a4h9ah8cxzkaw09rpzy369mkl70lfdk"
    }
  },
  block: {
    header: {
      version: "v1.0.0-rc.0",
      chain_id: "dev",
      height: "1",
      time: "2023-05-01T10:32:20.807541Z",
      num_txs: "0",
      total_txs: "0",
      app_version: "",
      last_block_id: {
        hash: null,
        parts: {
          total: "0",
          hash: null
        }
      },
      last_commit_hash: null,
      data_hash: null,
      validators_hash: "FnuBaDvDLg4FotGRcZAFvhLkEjkb+kNLaAZrAVhL5Aw=",
      next_validators_hash: "FnuBaDvDLg4FotGRcZAFvhLkEjkb+kNLaAZrAVhL5Aw=",
      consensus_hash: "uKhnXFmGUkxgQSJf17ogbYLNXDo3UEPwQvzddo4Vkuw=",
      app_hash: null,
      last_results_hash: null,
      proposer_address: "g1vsqzyy9a4h9ah8cxzkaw09rpzy369mkl70lfdk"
    },
    data: {
      txs: null
    },
    last_commit: {
      block_id: {
        hash: null,
        parts: {
          total: "0",
          hash: null
        }
      },
      precommits: null
    }
  }
}
*/
```

### getBlockResult

Fetches the block at the specific height, if any

#### Parameters

* `height` **number** the height for querying

Returns **Promise<BlockResult\>**

#### Usage

```ts
await provider.getBlockResult(1);
/*
{
  height: "1",
  results: {
    deliver_tx: null,
    end_block: {
      ResponseBase: {
        Error: null,
        Data: null,
        Events: null,
        Log: "",
        Info: ""
      },
      ValidatorUpdates: null,
      ConsensusParams: null,
      Events: null
    },
    begin_block: {
      ResponseBase: {
        Error: null,
        Data: null,
        Events: null,
        Log: "",
        Info: ""
      }
    }
  }
}
*/
```

### getBlockNumber

Fetches the latest block number from the chain

Returns **Promise<number\>**

#### Usage

```ts
await provider.getBlockNumber();
// 1300
```

## Network methods

### getNetwork

Fetches the network information

Returns **Promise<NetworkInfo\>**

#### Usage

```ts
await provider.getNetwork();
/*
{
  listening: true,
  listeners: [
    "Listener(@)"
  ],
  n_peers: "0",
  peers: []
}
*/
```

### getConsensusParams

Fetches the consensus params for the specific block height

#### Parameters

* `height` **number** the height for querying

Returns **Promise<ConsensusParams\>**

#### Usage

```ts
await provider.getConsensusParams(1);
/*
{
  block_height: "1",
  consensus_params: {
    Block: {
      MaxTxBytes: "1000000",
      MaxDataBytes: "2000000",
      MaxBlockBytes: "0",
      MaxGas: "10000000",
      TimeIotaMS: "100"
    },
    Validator: {
      PubKeyTypeURLs: [
        "/tm.PubKeyEd25519"
      ]
    }
  }
}
*/
```

### getStatus

Fetches the current node status

Returns **Promise<Status\>**

#### Usage

```ts
await provider.getStatus();
/*
{
  node_info: {
    version_set: [
      {
        Name: "abci",
        Version: "v1.0.0-rc.0",
        Optional: false
      },
      {
        Name: "app",
        Version: "",
        Optional: false
      },
      {
        Name: "bft",
        Version: "v1.0.0-rc.0",
        Optional: false
      },
      {
        Name: "blockchain",
        Version: "v1.0.0-rc.0",
        Optional: false
      },
      {
        Name: "p2p",
        Version: "v1.0.0-rc.0",
        Optional: false
      }
    ],
    net_address: "g1z0wa6rspsshkm2k7jlqvnjs8jdt4kvg4e9j640@0.0.0.0:26656",
    network: "dev",
    software: "",
    version: "v1.0.0-rc.0",
    channels: "QCAhIiMw",
    moniker: "voyager.lan",
    other: {
      tx_index: "off",
      rpc_address: "tcp://127.0.0.1:26657"
    }
  },
  sync_info: {
    latest_block_hash: "x5ewEBhf9+MGXbEFkUdOm3RsE40D+plUia2u0PuVfHs=",
    latest_app_hash: "7dB/+EmqLqEX2RkH2Zx+GcFo8c2vTs2ttW8urYyyFT4=",
    latest_block_height: "55",
    latest_block_time: "2023-05-06T11:28:35.643575Z",
    catching_up: false
  },
  validator_info: {
    address: "g1vsqzyy9a4h9ah8cxzkaw09rpzy369mkl70lfdk",
    pub_key: {
      "@type": "/tm.PubKeyEd25519",
      value: "X8ZS1DYu1eJ3HYnZ0OWk+0GgCdI7zA++kgWiprWMs3w="
    },
    voting_power: "0"
  }
}
*/
```

### getGasPrice

**NOTE: Not supported yet**

Fetches the current (recommended) average gas price

Returns **Promise<number\>**

### estimateGas

**NOTE: Not supported yet**

Estimates the gas limit for the transaction

#### Parameters

* `tx` **Tx** the transaction that needs estimating

Returns **Promise<number\>**

## Transaction methods

### sendTransaction

Sends the transaction to the node for committing and returns the transaction hash.
The transaction needs to be signed beforehand.

#### Parameters

* `tx` **string** the base64-encoded signed transaction

Returns **Promise<string\>**

#### Usage

```ts
await provider.sendTransaction('ZXhhbXBsZSBzaWduZWQgdHJhbnNhY3Rpb24');
// "dHggaGFzaA=="
```

### waitForTransaction

Waits for the transaction to be committed on the chain.
NOTE: This method will not take in the fromHeight parameter once
proper transaction indexing is added - the implementation should
simply try to fetch the transaction first to see if it's included in a block
before starting to wait for it; Until then, this method should be used
in the sequence:
get latest block -> send transaction -> waitForTransaction(block before send)

#### Parameters

* `hash` **string** The transaction hash
* `fromHeight` **number** The block height used to begin the search (optional, default `latest`)
* `timeout` **number** Optional wait timeout in MS (optional, default `15000`)

Returns **Promise<Tx\>**

#### Usage

```ts
await provider.waitForTransaction('ZXhhbXBsZSBzaWduZWQgdHJhbnNhY3Rpb24');
/*
{
   messages:[], // should be filled with the appropriate message type
   fee:{
      gasWanted: "100",
      gasFee: "1ugnot"
   },
   signatures:[
      {
         pubKey:[
            {
                type: "/tm.PubKeySecp256k1"
                value: "X8ZS1DYu1eJ3HYnZ0OWk+0GgCdI7zA++kgWiprWMs3w="
            }
         ],
         signature: "X8ZS1DYu1eJ3HYnZ0OWk+0GgCdI7zA++kgWiprWMs3w="
      }
   ],
   memo: "check out gno.land!"
}
*/
```
