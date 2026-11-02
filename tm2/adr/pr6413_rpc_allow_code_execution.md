# Refuse the queries that run code with rpc.allow_code_execution

## Status

Proposed in #6413.

## Context

Four `abci_query` paths run code on the node that serves them:
`.app/simulate` runs a whole transaction, and `vm/qeval`, `vm/qeval_json` and
`vm/qrender` evaluate an expression in a deployed package. A query pays no
fee, so a public node pays for that code in CPU time and memory, and no
setting turned it off. A proxy filtering URLs cannot see the path inside a
JSON-RPC body, a batch or a WebSocket frame.

## Decision

Add `rpc.allow_code_execution`, `true` by default. When it is `false`,
`Environment.ABCIQuery` refuses those paths before the query reaches the
application, with HTTP 403 on the URI transport and a JSON-RPC error on the
others. Every transport and the in-process client call that one method.

The match reads a path the way the application routes it: one leading `/`
dropped, everything from `?` cut, and every path below a listed one refused.

tm2 refuses `.app/simulate` itself, since `BaseApp` answers it. tm2 knows
nothing of the VM, so gno.land passes `vm.CodeExecutionQueryPaths` through
`node.WithCodeExecutionQueryPaths`, from `gnoland start` and from the
in-memory node.

## Alternatives considered

- Default `false` would stop free code execution at upgrade, and break
  gnoweb and the default `gnokey maketx` on every node that does not opt back
  in.
- The application could pass `.app/simulate` with its other paths, so tm2's
  RPC names no query path. A tm2 application that passes nothing would then
  have nothing refused.
- Refusing in the query handlers would also refuse code that calls the
  application directly. It needs two gates and the RPC config inside the app,
  and its ABCI error answers HTTP 200 rather than 403.
- Refusing `vm/qeval` alone leaves `.app/simulate` running a `MsgRun` that
  carries the same code.

## Consequences

The setting is node-local and nothing inside a block calls `ABCIQuery`, so
nodes with different values commit the same blocks. A `config.toml` written
before the key existed loads `true`. A node set to `false` still accepts
transactions through `broadcast_tx_*`, since `CheckTx` runs no message, while
`gnokey maketx` needs `-simulate skip` and gnoweb cannot render through it.
A VM query added later that builds a machine is refused only once it joins
`vm.CodeExecutionQueryPaths`.

Tests send each spelling to `ABCIQuery`, send a refused and a served call
over a URI GET, a JSON-RPC batch and a WebSocket, boot an in-memory node and
`gnoland start` with the setting `false`, and load a config file without the
key.
