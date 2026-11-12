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

The application says which paths run code, through the optional interface
`abci.CodeExecutionQuerier`. The node finds it on the local application, the
way it finds `abci.DeferredCommitter`, and hands `QueryRunsCode` to the RPC
environment, so no node constructor has to remember an option. `BaseApp`
answers with the parser its `Query` routes by: `.app/simulate` is its own, and
a custom route asks its handler when the handler implements the same
interface. The VM handler answers with the endpoint parser its `Query`
switches on, and lists `qrender`, `qeval` and `qeval_json` beside that switch.
A path is refused exactly when the application would run code for it.

## Alternatives considered

- Default `false` would stop free code execution at upgrade, and break
  gnoweb and the default `gnokey maketx` on every node that does not opt back
  in.
- The RPC could match path strings itself, from a list each node constructor
  passes as an option. That copies the application's routing into tm2's RPC,
  where a change to either drifts silently, and a constructor that forgets the
  option leaves the VM paths open.
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
`gnokey maketx` needs `-simulate skip` and gnoweb cannot render or evaluate
through it. A VM query added later that builds a machine is refused only once it joins the VM handler's
`QueryRunsCode`. An application that does not implement
`abci.CodeExecutionQuerier` has nothing refused.

Tests check each spelling against `BaseApp` and the VM handler, send a refused
and a served call over a URI GET, a JSON-RPC batch and a WebSocket, boot an
in-memory node with the setting `false` and `gnoland start` with each value,
and load a config file without the key.
