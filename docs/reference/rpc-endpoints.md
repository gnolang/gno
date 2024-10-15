---
id: rpc-endpoints
---

# Gno RPC Endpoints

For network configurations, view the [network configuration page](./network-config.md).
## Common Parameters

#### Response

| Name            | Type   | Description                       |
| --------------- | ------ | --------------------------------- |
| `jsonrpc`       | String | The RPC version.                  |
| `id`            | String | The response ID.                  |
| `result`        | Object | (upon success) The result object. |
| `error`         | Object | (upon failure) The error object.  |
| `error.code`    | Number | The error code.                   |
| `error.message` | String | The error message.                |
| `error.data`    | String | The error data.                   |

## Health Check

Call with the `/health` path when verifying that the node is running.

#### Response

| Name      | Type   | Description      |
| --------- | ------ | ---------------- |
| `jsonrpc` | String | The RPC version. |
| `id`      | String | The response ID. |
| `result`  | Object | {}               |

## Check Node Server Status

Call with the `/status` path to check the information from a node.

#### Response

| Name      | Type             | Description                           |
| --------- | ---------------- | ------------------------------------- |
| `jsonrpc` | String           | The RPC version.                      |
| `id`      | String           | The response ID.                      |
| `result`  | \[Status Result] | The result of the node server status. |

#### Status Result

| Name             | Type   | Description                         |
| ---------------- | ------ | ----------------------------------- |
| `node_info`      | Object | General information about the node. |
| `sync_info`      | Object | The sync information.               |
| `validator_info` | Object | The validator information.          |

## Get Network Information

Call with the `/net_info` path to check the network information from the node.

#### Response

| Name      | Type              | Description              |
| --------- | ----------------- | ------------------------ |
| `jsonrpc` | String            | The RPC version.         |
| `id`      | String            | The response ID.         |
| `result`  | \[NetInfo Result] | The network information. |

#### NetInfo Result

| Name        | Type       | Description        |
| ----------- | ---------- | ------------------ |
| `listening` | Boolean    | Enables listening. |
| `listeners` | String \[] | List of listeners. |
| `n_peers`   | String     | Number of peers.   |
| `peers`     | String \[] | List of peers.     |

## Get Genesis Block Information

Call with the `/genesis` path to retrieve information about the Genesis block from the node.

#### Response

| name      | Type   | Description                    |
| --------- | ------ | ------------------------------ |
| `jsonrpc` | String | The RPC version.               |
| `id`      | String | The response ID.               |
| `result`  | Object | The Genesis block information. |

## Get Consensus Parameters

Call with the /consensus\_params path to check the consensus algorithm parameters at the specified height.

#### Parameters

| Name     | Description       |
| -------- | ----------------- |
| `height` | The block height. |

#### Response

| Name      | Type                       | Description                          |
| --------- | -------------------------- | ------------------------------------ |
| `jsonrpc` | String                     | The RPC Version.                     |
| `id`      | String                     | The response ID.                     |
| `result`  | \[Consensus Params Result] | The consensus parameter information. |

#### Consensus Params Result

| Name                          | Type   | Description                |
| ----------------------------- | ------ | -------------------------- |
| `block_height`                | String | The block height.          |
| `consensus_params`            | Object | The parameter information. |
| `consensus_params.Block`      | Object | The block parameters.      |
| `consensus_params.Validator` | Object | The validator parameters.  |

## Get Consensus State

Call with the `/consensus_state` to get the consensus state of the Gnoland blockchain

#### Response

| Name    | Type                        | Description                      |
| ------- | --------------------------- | -------------------------------- |
| jsonrpc | String                      | The RPC version.                 |
| id      | String                      | The response ID.                 |
| result  | \[Consensus State Response] | The consensus state information. |

#### Consensus State Response

| Name                              | Type   | Description                      |
| --------------------------------- | ------ | -------------------------------- |
| `round_state`                     | Object | The consensus state object.      |
| `round_state.height/round/step`   | String | The block height / round / step. |
| `round_state.start_time`          | String | The round start time.            |
| `round_state.proposal_block_hash` | String | The proposal block hash.         |
| `round_state.locked_block_hash`   | String | The locked block hash.           |
| `round_state.valid_block_hash`    | String | The valid block hash.            |
| `round_state.height_vote_set`     | Object | -                                |

## Get Commit

Call with the `/commit` path to retrieve commit information at the specified block height.

#### Parameters

| Name     | Description       |
| -------- | ----------------- |
| `height` | The block height. |

#### Response

| Name      | Type             | Description             |
| --------- | ---------------- | ----------------------- |
| `jsonrpc` | String           | The RPC version.        |
| `id`      | String           | The response ID.        |
| `result`  | \[Commit Result] | The commit information. |

#### Commit Result

| Name           | Type    | Description               |
| -------------- | ------- | ------------------------- |
| signed\_header | Object  | The signed header object. |
| canonical      | Boolean | Returns commit state.     |

## Get Block Information

Call with the `/block` path to retrieve block information at the specified height.

#### Parameters

| Name     | Description       |
| -------- | ----------------- |
| `height` | The block height. |

#### Response

| Name      | Type            | Description             |
| --------- | --------------- | ----------------------- |
| `jsonrpc` | String          | The RPC version.        |
| `id`      | String          | The response ID.        |
| `result`  | \[Block Result] | The commit information. |

#### Block Result

| Name         | Type   | Description            |
| ------------ | ------ | ---------------------- |
| `block_meta` | Object | The block metadata.    |
| `block`      | Object | The block information. |

## Get Block Results

Call with the `/block_results` path to retrieve block processing information at the specified height.

#### Parameters

| Name     | Description       |
| -------- | ----------------- |
| `height` | The block height. |

#### Response

| Name      | Type            | Description        |
| --------- | --------------- | ------------------ |
| `jsonrpc` | String          | The RPC version.   |
| `id`      | String          | The response ID.   |
| `result`  | \[Block Result] | The result object. |

#### Block Result

| Name      | Type                     | Description                           |
| --------- | ------------------------ | ------------------------------------- |
| `height`  | Object                   | The block height.                     |
| `results` | \[Block Result Info] \[] | The list of block processing results. |

#### Block Result Info

| Name                        | Type       | Description                      |
| --------------------------- | ---------- | -------------------------------- |
| `deliver_tx`                | Object \[] | The list of transaction results. |
| `deliver_tx[].ResponseBase` | Object     | The transaction response object. |
| `deliver_tx[].GasWanted`    | String     | Maximum amount of gas to use.    |
| `deliver_tx[].GasUsed`      | String     | Actual gas used.                 |
| `begin_block`               | Object     | Previous block information.      |
| `end_block`                 | Object     | Next block information.          |

## Get Block List

Call with the `/blockchain` path to retrieve information about blocks within a specified range.

#### Parameters

| Name        | Description               |
| ----------- | ------------------------- |
| `minHeight` | The minimum block height. |
| `maxHeight` | The maximum block height. |

#### Response

| Name      | Type                 | Description        |
| --------- | -------------------- | ------------------ |
| `jsonrpc` | String               | The RPC version.   |
| `id`      | String               | The response ID.   |
| `result`  | \[Blockchain Result] | The result object. |

#### Blockchain Result

| Name          | Type       | Description                 |
| ------------- | ---------- | --------------------------- |
| `last_height` | String     | The latest block height.    |
| `block_meta`  | Object \[] | The list of block metadata. |

## Get a No. of Unconfirmed Transactions

Call with the `/num_unconfirmed_txs` path to get data about unconfirmed transactions.

#### Response

| Name      | Type                          | Description        |
| --------- | ----------------------------- | ------------------ |
| `jsonrpc` | String                        | The RPC version.   |
| `id`      | String                        | The response ID.   |
| `result`  | \[Num Unconfirmed Txs Result] | The result object. |

#### Num Unconfirmed Txs Result

| Name          | Type   | Description                 |
| ------------- | ------ | --------------------------- |
| `n_txs`       | String | The number of transactions. |
| `total`       | String | The total number.           |
| `total_bytes` | String | Total bytes.                |
| `txs`         | null   | -                           |

## Get a List of Unconfirmed Transactions

Call with the `/unconfirmed_txs` path to get a list of unconfirmed transactions.

#### Parameters

| Name    | Description                             |
| ------- | --------------------------------------- |
| `limit` | The maximum transaction numbers to get. |

#### Response

| Name      | Type                      | Description        |
| --------- | ------------------------- | ------------------ |
| `jsonrpc` | String                    | The RPC version.   |
| `id`      | String                    | The response ID.   |
| `result`  | \[Unconfirmed Txs Result] | The result object. |

#### Unconfirmed Txs Result

| Name          | Type       | Description                         |
| ------------- | ---------- | ----------------------------------- |
| `n_txs`       | String     | The number of transactions.         |
| `total`       | String     | The total number.                   |
| `total_bytes` | String     | Total bytes.                        |
| `txs`         | Object \[] | A list of unconfirmed transactions. |

## Get a List of Validators

Call with the `/validators` path to get a list of validators at a specific height.

#### Parameters

| Name     | Description                               |
| -------- | ----------------------------------------- |
| `height` | The block height (default: newest block). |

#### Response

| Name      | Type                 | Description        |
| --------- | -------------------- | ------------------ |
| `jsonrpc` | String               | The RPC version.   |
| `id`      | String               | The response ID.   |
| `result`  | \[Validators Result] | The result object. |

#### Validators Result

| Name           | Type             | Description             |
| -------------- | ---------------- | ----------------------- |
| `block_height` | Object           | The block height.       |
| `validators`   | \[Validator] \[] | The list of validators. |

#### Validator

| Name                | Type       | Description                              |
| ------------------- | ---------- | ---------------------------------------- |
| `address`           | String     | The address of the validator.            |
| `pub_key`           | Object \[] | The public key object of the validator.  |
| `pub_key.@type`     | String     | The type of validator's public key.      |
| `pub_key.value`     | String     | The value of the validator's public key. |
| `voting_power`      | String     | Voting power of the validator.           |
| `proposer_priority` | String     | The priority of the proposer.            |

## Broadcast a Transaction - Asynchronous

Call with the `/broadcast_tx_async` path to create and broadcast a transaction without waiting for the transaction response.

#### Parameters

| Name | Description                                 |
| ---- | ------------------------------------------- |
| `tx` | The value of the signed transaction binary. |

#### Response

| Name      | Type                  | Description        |
| --------- | --------------------- | ------------------ |
| `jsonrpc` | String                | The RPC version.   |
| `id`      | String                | The response ID.   |
| `result`  | \[Transaction Result] | The result object. |

#### Transaction Result

| Name  | Type   | Description                  |
| ----- | ------ | ---------------------------- |
| hash  | String | The transaction hash.        |
| data  | Object | The transaction data object. |
| error | Object | The error object.            |
| log   | String | The log information.         |

## Broadcast a Transaction - Synchronous

Call with the `/broadcast_tx_sync` path to create and broadcast a transaction, then wait for the transaction response.

#### Parameters

| Name | Description                                 |
| ---- | ------------------------------------------- |
| `tx` | The value of the signed transaction binary. |

#### Response

| Name      | Type                  | Description        |
| --------- | --------------------- | ------------------ |
| `jsonrpc` | String                | The RPC version.   |
| `id`      | String                | The response ID.   |
| `result`  | \[Transaction Result] | The result object. |

#### Transaction Result

| Name  | Type   | Description                  |
| ----- | ------ | ---------------------------- |
| hash  | String | The transaction hash.        |
| data  | Object | The transaction data object. |
| error | Object | The error object.            |
| log   | String | The log information.         |

## (NOT RECOMMENDED) Broadcast Transaction and Get Commit Information

Call with the `/broadcast_tx_commit` path to create and broadcast a transaction, then wait for the transaction response and the commit response.

#### Parameters

| Name | Description                                 |
| ---- | ------------------------------------------- |
| `tx` | The value of the signed transaction binary. |

#### Response

| Name      | Type                         | Description        |
| --------- | ---------------------------- | ------------------ |
| `jsonrpc` | String                       | The RPC version.   |
| `id`      | String                       | The response ID.   |
| `result`  | \[Transaction Commit Result] | The result object. |

#### Transaction Commit Result

| Name         | Type   | Description                                                 |
| ------------ | ------ | ----------------------------------------------------------- |
| `height`     | String | The height of the block when the transaction was committed. |
| hash         | String | The transaction hash.                                       |
| `deliver_tx` | Object | The delivered transaction information.                      |
| `check_tx`   | Object | The committed transaction information.                      |

## ABCI

### Get ABCI Information

Call with the `/abci_info` path to get the latest information about the ABCI.

#### Response

| Name      | Type                | Description             |
| --------- | ------------------- | ----------------------- |
| `jsonrpc` | String              | The RPC version.        |
| `id`      | String              | The response ID.        |
| `result`  | \[ABCI Info Result] | The commit information. |

#### ABCI Info Result

| Name                        | Type             | Description                |
| --------------------------- | ---------------- | -------------------------- |
| `response`                  | Object           | The metadata of the block. |
| `response.ResponseBase`     | \[ABCI Response] | The ABCI response data.    |
| `response.ABCIVersion`      | String           | The ABCI version.          |
| `response.AppVersion`       | String           | The app version.           |
| `response.LastBlockHeight`  | String           | The latest block height.   |
| `response.LastBlockAppHash` | String           | The latest block hash.     |

#### ABCI Response

| Name   | Type       | Description                       |
| ------ | ---------- | --------------------------------- |
| Data   | String     | The Base64-encoded response data. |
| Error  | Object     | The ABCI response error object.   |
| Events | Object \[] | The list of event objects.        |
| Log    | String     | The ABCI response log.            |
| Info   | String     | The ABCI response information.    |

### Get ABCI Query

Call with the `/abci_query` to get information via the ABCI Query.

#### Query

| Name                      | Description                                                        |
| ------------------------- | ------------------------------------------------------------------ |
| `auth/accounts/{ADDRESS}` | Returns the account information.                                   |
| `bank/balances/{ADDRESS}` | Returns the balance information about the account.                 |
| `vm/qfuncs`               | Returns public facing function signatures as JSON.                 |
| `vm/qfile`                | Returns the file bytes, or list of files if directory.             |
| `vm/qrender`              | Calls `.Render(<path>)` in readonly mode.                          |
| `vm/qeval`                | Evaluates any expression in readonly mode and returns the results. |
| `vm/store`                | (not yet supported) Fetches items from the store.                  |
| `vm/package`              | (not yet supported) Fetches a package's files.                     |

#### Parameters

| Name                | Description                                      |
| ------------------- | ------------------------------------------------ |
| `path`              | The query path.                                  |
| `data`              | The data from the query path.                    |
| (optional) `height` | The block height (default: latest block height). |
| (optional) `prove`  | The validation status.                           |

#### Response

| Name      | Type                 | Description             |
| --------- | -------------------- | ----------------------- |
| `jsonrpc` | String               | The RPC version.        |
| `id`      | String               | The response ID.        |
| `result`  | \[ABCI Query Result] | The commit information. |

#### ABCI Query Result

| Name                    | Type             | Description                |
| ----------------------- | ---------------- | -------------------------- |
| `response`              | Object           | The metadata of the block. |
| `response.ResponseBase` | \[ABCI Response] | The ABCI response data.    |
| `response.Key`          | String           | The key.                   |
| `response.Value`        | String           | The value.                 |
| `response.Proof`        | String           | The validation ID.         |
| `response.Height`       | String           | The block height.          |

#### ABCI Response

| Name   | Type       | Description                       |
| ------ | ---------- | --------------------------------- |
| Data   | String     | The Base64-encoded response data. |
| Error  | Object     | The ABCI response error object.   |
| Events | Object \[] | The list of event objects.        |
| Log    | String     | The ABCI response log.            |
| Info   | String     | The ABCI response information.    |
