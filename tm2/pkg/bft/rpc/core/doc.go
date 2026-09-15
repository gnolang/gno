/*
Package core implements the JSON-RPC methods a Tendermint2 node exposes.

The route map in [Environment.Routes] is the authoritative list of methods and
their parameters. For the user-facing reference — encoding rules and limits —
see docs/resources/rpc-endpoints.md.

# Transports

The same methods are served three ways, all mounted by [rpc.RegisterRPCFuncs]
and [github.com/gnolang/gno/tm2/pkg/bft/node]:

  - URI over HTTP, as GET /<method>?<arg>=<value>
  - JSON-RPC over HTTP, as POST / with a request object
  - WebSocket, at /websocket, serving the same methods

Both JSON-RPC transports also accept an array of request objects as a batch.

A request to / with an empty body returns an HTML index of the methods
available on that node.

# Configuration

Parameters live under the rpc table of the node's config.toml, inside the data
directory given by gnoland's --data-dir, and are edited there or with
"gnoland config set rpc.<key> <value>". The default listen address is
tcp://127.0.0.1:26657.
The unsafe_* methods are registered only when rpc.unsafe is true. Two of
them pass a caller-supplied filename straight to os.Create, so a node running
with rpc.unsafe on a reachable address lets any caller overwrite files as the
node user.

# Arguments

Byte-array arguments may be passed as base64, or — on the URI transport only —
as a 0x-prefixed hex string such as 0x616263. The 0x form is decoded in
httpParamsToArgs; the JSON-RPC transport accepts base64 exclusively.

String arguments are safest quoted, as path="auth/accounts/g1...". An unquoted
value is wrapped for the caller unless it already parses as JSON. A bare true
or an out-of-range number then reaches amino as raw JSON and fails to unmarshal
into a string; a bare null is worse, since it unmarshals silently to "".

The JSON-RPC envelope is ordinary JSON, but the result is marshalled with
Amino JSON, which encodes byte arrays as base64 and 64-bit integers as quoted
strings. Types with their own MarshalAmino, addresses among them, override
that.
*/
package core
