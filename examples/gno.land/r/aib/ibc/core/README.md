# r/aib/ibc

Because most of the functions in this realm take complex args, it is required
to call them using `MsgRun` (`maketx run` with the CLI) instead of the more
commonly used `MsgCall`.

Here is an exemple of the command:

```
$ gnokey maketx run -gas-fee 1000000ugnot -gas-wanted 90000000 \
    -broadcast -chainid "dev" -remote "tcp://127.0.0.1:26657" \
    ADDRESS run.gno
```

`run.gno` content depends on the called function, see the following sections
for examples.

## CreateClient

See [`zz_create_client_example_filetest.gno`](./zz_create_client_example_filetest.gno)

Emitted event:
```json
{
  "type": "create_client",
  "attrs": [
    {
      "key": "client_id",
      "value": "07-tendermint-1"
    },
    {
      "key": "client_type",
      "value": "07-tendermint"
    },
    {
      "key": "consensus_heights",
      "value": "2/2"
    }
  ],
  "pkg_path": "gno.land/r/aib/ibc/core"
}
```

## UpdateClient

See [`zz_update_client_example_filetest.gno`](./zz_update_client_example_filetest.gno)

Emitted event:
```json
{
  "type": "update_client",
  "attrs": [
    {
      "key": "client_id",
      "value": "07-tendermint-1"
    },
    {
      "key": "client_type",
      "value": "07-tendermint"
    },
    {
      "key": "consensus_heights",
      "value": "2/5"
    }
  ],
  "pkg_path": "gno.land/r/aib/ibc/core"
}
```
