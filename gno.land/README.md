# Gno.land

Gno.land is a layer-1 blockchain that integrates various cutting-edge technologies, including [Tendermint2](../tm2), [GnoVM](../gnovm), Proof-of-Contributions consensus mechanism, on-chain governance through a new DAO framework with support for sub-DAOs, and a unique licensing model that allows open-source code to be monetized by default.

## Getting started

Use [`gnokey`](./cmd/gnokey) to interact with Gno.land's testnets, localnet, and upcoming mainnet.

For localnet setup, use [`gnoland`](./cmd/gnoland).

To add a web interface and faucet to your localnet, use [`gnoweb`](./cmd/gnoweb) and [`gnofaucet`](../contribs/gnofaucet).

## Interchain

Gno.land aims to offer security, high-quality contract libraries, and scalability to other Gnolang chains, while also prioritizing interoperability with existing and emerging chains.

Post mainnet launch, Gno.land aims to integrate IBCv1 to connect with existing Cosmos chains and implement ICS1 for security through the existing chains.
Afterwards, the platform plans to improve IBC by adding new capabilities for interchain smart-contracts.

## Under the hood

* [Tendermint2](../tm2): a secure and stable consensus engine
* [GnoVM](../gnovm): a Virtual-Machine that provides transparency and security
* Proof-of-Contributions: a new consensus mechanism secured by contributors
* On-chain governance: managed by a new DAO framework with support for sub-DAOs
* Licensing model: a unique approach that allows open-source code to be monetized by default
* Interoperability and shared security: IBCv1, IBCx, ICS1, ICSx
