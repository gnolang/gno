---
id: tm2-js-ws-provider
---

# WebSocket Provider

Provider based on WS JSON-RPC requests.

### new WSProvider

Creates a new instance of the WebSocket Provider

#### Parameters

* `baseURL` **string** the WS URL of the node
* `requestTimeout` **number** the timeout for the WS request (in MS)

#### Usage

```ts
new WSProvider('ws://staging.gno.land:36657/ws');
// provider with WS connection is created
```

### closeConnection

Closes the WS connection. Required when done working
with the WS provider

#### Usage

```ts
const wsProvider = new WSProvider('ws://staging.gno.land:36657/ws');

wsProvider.closeConnection();
// WS connection is now closed
```

### sendRequest

Sends a request to the WS connection, and resolves
upon receiving the response

#### Parameters

* `request` **RPCRequest** the RPC request

Returns **Promise<RPCResponse<Result\>>**

#### Usage

```ts
const request: RPCRequest = // ...

const wsProvider = new WSProvider('ws://staging.gno.land:36657/ws');

wsProvider.sendRequest<Result>(request);
// request is sent over the open WS connection
```

### parseResponse

Parses the result from the response

#### Parameters

* `response` **RPCResponse<Result\>** the response to be parsed

Returns **Result**

#### Usage

```ts
const response: RPCResponse = // ...

const wsProvider = new WSProvider('ws://staging.gno.land:36657/ws');

wsProvider.parseResponse<Result>(response);
// response is parsed
```

### waitForOpenConnection

Waits for the WS connection to be established

Returns **Promise<null\>**

#### Usage

```ts
const wsProvider = new WSProvider('ws://staging.gno.land:36657/ws');

await wsProvider.waitForOpenConnection()
// status of the connection is: CONNECTED
```
