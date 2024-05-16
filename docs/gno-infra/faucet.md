---
id: faucet
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

## Premining funds to an address

Before setting up the faucet, we need to make sure that the address will be used
to serve the funds contain enough testnet funds. 

In your monorepo clone, visit the `genesis_balances.txt` file in the 
`gno.land/genesis` folder. This file contains a list of addresses and their
initial balances upon chain initialization. The file follows the pattern below
for premining specific amounts of GNOT to an address.

```
g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5=10000000000000ugnot # test1
g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj=10000000000000ugnot # test2
```

Add the address you plan to use for your faucet in the same format:

```
<address>=<amount>ugnot
```

After this, you can spin up your chain and run the following command to check
that the address indeed contains the intended balances:

```bash
gnokey query bank/balances/<address> --remote <node_rpc_listener_address>
```

Running this command will output something like the following:

```bash
height: 0
data: "10000000000000ugnot"
```

Now this address is ready to be used for the faucet.

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

We are now ready to run the faucet.

## Running the faucet

To run the faucet, you will need to use the mnemonic phrase from which your 
faucet address is derived from. Then, simply run the following command:

```bash
> ./build/faucet --mnemonic "<faucet_account_mnemonic>"

time=2024-05-16T11:25:36.012+02:00 level=INFO msg="faucet started at [::]:8545"
```

The faucet should be running on `localhost:8545`, and is connected to the locally
running `gnoland` instance. By default, `gnoland`'s rpc listener address is matched
in the `--remote` flag in the faucet. If your node is listening on a separate
address, make sure to match it accordingly when running the faucet.

For further configuration options, such as the maximum drip amount for the faucet,
check out the [faucet reference page](../gno-tooling/cli/faucet/faucet.md).

## Making faucet requests

The faucet takes in standard HTTP post requests with JSON data. The basic request
format is the following:

```json
{
  "To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh"
}
```

You can test this out buy running the following `curl` command:
```bash
curl --location --request POST 'http://localhost:8545' --header 'Content-Type: application/json' --data '{"To": "g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh"}'
```

## Conclusion

Congratulations, you've successfully set up a faucet for your Gno.land network.
















