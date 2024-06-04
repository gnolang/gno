---
id: setting-up-a-faucet
---

# Setting up a faucet for your Gno network

In this tutorial, we will cover how to run a local native currency faucet that 
works seamlessly with a Gno node. Using the faucet, any address can get a hold
of testnet GNOTs.

## Prerequisites
- Git
- Go 1.21+
- Make (for running Makefiles)
- `gnoland` & `gnokey` installed
- A Gno.land keypair generated using [`gnokey`](../gno-tooling/cli/gnokey.md) 
containing funds for the faucet to serve

## Cloning the repo

To get started with setting up the faucet, visit the 
[faucet repo](https://github.com/gnolang/faucet) and clone it:

```bash
git clone git@github.com:gnolang/faucet.git
```

After going into the cloned folder, you can build out the faucet binary:
```bash
make build
```

We are now ready to configure & run the faucet.

## Configuring the faucet

By running the `generate` command from the faucet binary, you will be able to generate
a `config.toml` file.

```bash
./build/faucet generate
```

In the `config.toml` file, you will be able to configure a few parameters:
- ChainID of the node to connect to
- Faucet listener address
- Mnemonic phrase to use for generating the account(s) to serve funds from
- The number of accounts to generate from the mnemonic
- The maximum drip amount for the faucet
- CORS configuration of the faucet

The default config file looks like this:
```yaml
chain_id = "dev"
listen_address = "0.0.0.0:8545"
mnemonic = "<your_mnemonic_phrase>"
num_accounts = 1
send_amount = "1000000ugnot"

[cors_config]
  cors_allowed_headers = ["Origin", "Accept", "Content-Type", "X-Requested-With", "X-Server-Time"]
  cors_allowed_methods = ["HEAD", "GET", "POST", "OPTIONS"]
  cors_allowed_origins = ["*"]
``` 

After inputting the mnemonic phrase from which your faucet address is derived 
from, you are ready to run the faucet.

## Running the faucet

To run the faucet, simply run the following command: 

```bash
> ./build/faucet serve --faucet-config <path_to_config.toml>

time=2024-05-16T11:25:36.012+02:00 level=INFO msg="faucet started at [::]:8545"
```

The faucet should be running on `localhost:8545`, and is connected to the locally
running `gnoland` instance. By default, `gnoland`'s rpc listener address is matched
in the `--remote` flag in the faucet. If your node is listening on a separate
address, make sure to match it accordingly when running the faucet.

## Making faucet requests

The faucet takes in standard HTTP post requests with JSON data. The basic request
format is the following:

```json
{
  "To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh"
}
```

By default, this will send the maximum allowed amount to the address, as specified 
in the `config.toml` file under the `send_amount` field. A request can also be made with a 
specific amount of `ugnot`:

```json
{
  "To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh",
  "Amount": "100ugnot"
}
```

You can test the requests by running the following `curl` command, and inputting
the request under the `--data` field:
```bash
curl --location --request POST 'http://localhost:8545' --header 'Content-Type: application/json' --data '{"To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh","Amount": "100ugnot"}'
```

If the request is successful, you should get a response similar to the following:
```bash
{"result":"successfully executed faucet transfer"}
```

The faucet also supports batch requests, so a request such as the following is 
also valid:
```json
[
  {
    "To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh",
    "Amount": "100ugnot"
  },
  {
    "To": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj",
    "Amount": "200ugnot"
  }
]
```

Sending this to the faucet will receive the following response:

```json
[
    {
        "result": "successfully executed faucet transfer"
    },
    {
        "result": "successfully executed faucet transfer"
    }
]
```

## Faucet errors

Below are errors you may run into when setting up or using the faucet.

### During setup

When setting up the faucet, you can run into the following errors:
- If the faucet listen address is invalid or is taken - `invalid listen address`
- If the chain ID the faucet connects to is invalid - `invalid chain ID`
- If the send amount defined is invalid - `invalid send amount`
- If the mnemonic used for the faucet is invalid - `invalid mnemonic`
- If the number of accounts to derive from the mnemonic is less than zero -
`invalid number of faucet accounts`

### During requests

When requesting a drip from the faucet, you can face the following errors:
- If the address provided is empty or has an invalid checksum - `invalid beneficiary address`
- If the amount requested is empty, not in the `<amount>ugnot` format, or is larger
than `send_amount` defined in the faucet configuration

## Extending the faucet

This faucet can be used as a library and can be extended with middleware and other
layers of security. To use the faucet in your project, run the following command:

```
go get github.com/gnolang/faucet
```

To then use the faucet in-code, you can set up your project the following way:

```go
package main

import (
	// ...
	"context"

	"github.com/gnolang/faucet/client/http"
	"github.com/gnolang/faucet/estimate/static"
)

func main() {
	// Create the faucet
	f, err := NewFaucet(
		static.New(...), // gas estimator
		http.NewClient(...), // remote address 
        )

	// The faucet is controlled through a top-level context
	ctx, cancelFn := context.WithCancel(context.Background())

	// Start the faucet
	go f.Serve(ctx)

	// Close the faucet
	cancelFn()
}
```

To see an example of how the faucet can be extended, check out 
[`gnofaucet`](https://github.com/gnolang/gno/tree/master/contribs/gnofaucet).

## Conclusion

That's it 🎉

You have successfully set up a GNOT faucet on for a local Gno.land chain!
Read more about the faucet on the [`faucet`](https://github.com/gnolang/faucet) repo.
