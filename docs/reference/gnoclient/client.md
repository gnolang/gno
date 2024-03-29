---
id: client
---

# Client

## type [Client](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client.go#L8-L11>)

`Client` provides an interface for interacting with the blockchain. It is the main
struct of the `gnoclient` package, exposing all APIs used to communicate with a 
Gno.land chain.

```go
type Client struct {
    Signer    Signer           // Signer for transaction authentication
    RPCClient rpcclient.Client // RPC client for blockchain communication
}
```

### func \(\*Client\) [Call](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L56>)

```go
func (c *Client) Call(cfg BaseTxCfg, msgs ...MsgCall) (*ctypes.ResultBroadcastTxCommit, error)
```

`Call` executes a one or more [MsgCall](#type-msgcall) calls on the blockchain.

### func \(\*Client\) [Send](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L176>)

```go
func (c *Client) Send(cfg BaseTxCfg, msgs ...MsgSend) (*ctypes.ResultBroadcastTxCommit, error)
```

`Send` executes one or more [MsgSend](#type-msgsend) calls on the blockchain.

### func \(\*Client\) [Run](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L112>)

```go
func (c *Client) Run(cfg BaseTxCfg, msgs ...MsgRun) (*ctypes.ResultBroadcastTxCommit, error)
```

`Run` executes a one or more MsgRun calls on the blockchain.

### func \(*Client\) [QEval](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L108>)

```go
func (c *Client) QEval(pkgPath string, expression string) (string, *ctypes.ResultABCIQuery, error)
```

`QEval` evaluates the given expression with the realm code at `pkgPath`.
The `pkgPath` should include the prefix like `gno.land/`. The expression is 
usually a function call like `Render("")`.

### func \(*Client\) [Query](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L22>)

```go
func (c *Client) Query(cfg QueryCfg) (*ctypes.ResultABCIQuery, error)
```

`Query` performs a generic query on the blockchain.

### func \(*Client\) [QueryAccount](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L39>)

```go
func (c *Client) QueryAccount(addr crypto.Address) (*std.BaseAccount, *ctypes.ResultABCIQuery, error)
```

`QueryAccount` retrieves account information for a given address.

### func \(*Client\) [QueryAppVersion](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L65>)

```go
func (c *Client) QueryAppVersion() (string, *ctypes.ResultABCIQuery, error)
```

`QueryAppVersion` retrieves information about the Gno.land app version.

### func \(*Client\) [Render](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L85>)

```go
func (c *Client) Render(pkgPath string, args string) (string, *ctypes.ResultABCIQuery, error)
```

`Render` calls the Render function for pkgPath with optional args. The `pkgPath`
should include the prefix like `gno.land/`. This is similar to using a browser
URL `<testnet>/<pkgPath>:<args>` where `<pkgPath>` doesn't have the prefix like
`gno.land/`.

## type [BaseTxCfg](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L27-L33>)

`BaseTxCfg` defines the base transaction configuration, shared by all message
types.

```go
type BaseTxCfg struct {
    GasFee         string // Gas fee
    GasWanted      int64  // Gas wanted
    AccountNumber  uint64 // Account number
    SequenceNumber uint64 // Sequence number
    Memo           string // Memo
}
```

## type [MsgCall](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L36-L41>)

`MsgCall` \- syntax sugar for `vm.MsgCall`.

```go
type MsgCall struct {
    PkgPath  string   // Package path
    FuncName string   // Function name
    Args     []string // Function arguments
    Send     string   // Send amount
}
```

## type [MsgRun](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L50-L53>)

`MsgRun` \- syntax sugar for `vm.MsgRun`.

```go
type MsgRun struct {
    Package *std.MemPackage // Package to run
    Send    string          // Send amount
}
```

## type [MsgSend](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_txs.go#L44-L47>)

`MsgSend` \- syntax sugar for `bank.MsgSend`.

```go
type MsgSend struct {
    ToAddress crypto.Address // Send to address
    Send      string         // Send amount
}
```

## type [QueryCfg](<https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/client_queries.go#L15-L19>)

`QueryCfg` contains configuration options for performing ABCI queries.

```go
type QueryCfg struct {
    Path                       string // Query path
    Data                       []byte // Query data
    rpcclient.ABCIQueryOptions        // ABCI query options
}
```