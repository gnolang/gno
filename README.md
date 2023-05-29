# Gno

At first, there was Bitcoin, out of entropy soup of the greater All.
Then, there was Ethereum, which was created in the likeness of Bitcoin,
but made Turing complete.

Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
Then came Gno upon Cosmos and there spring forth Gnoland,
simulated by the Gnomes of the Greater Resistance.

## Discover

* [examples](./examples) - smart-contract examples and guides for new Gno developers.
* [gnovm](./gnovm) - GnoVM and Gnolang.
* [gno.land](./gno.land) - Gno.land blockchain and tools.
* [tm2](./tm2) - Tendermint2.

## Getting started

Start your journey with Gno.land by:
- using the [`gnoweb`](./gno.land/cmd/gnoweb) interface on the [latest testnet (test3.gno.land)](https://test3.gno.land/),
- sending transactions with [`gnokey`](./gno.land/cmd/gnokey),
- writing smart-contracts with [`gno` (ex `gnodev`)](./gnovm/cmd/gno).

Also, see the [quickstart guide](https://test3.gno.land/r/demo/boards:testboard/5).

## Contact

 * Discord: https://discord.gg/YFtMjWwUN7 <-- join now
 * Gnoland: https://gno.land/r/demo/boards:testboard
 * Telegram: https://t.me/gnoland
 * Twitter: https://twitter.com/_gnoland

<details><summary>Short doc about all the commands</summary>

  User commands:

  * [gnokey](./gno.land/cmd/gnokey) - key manipulation, also general interaction with gnoland
  * [gnoland](./gno.land/cmd/gnoland) - runs the blockchain node
  * [gnoweb](./gno.land/cmd/gnoweb) - serves gno website, along with user-defined content
  * [logos](./misc/logos) - intended to be used as a browser

  Developer commands:

  * [gno](./gnovm/cmd/gno) - handy tool for developing gno packages & realms
  * [gnotxsync](./tm2/cmd/gnotxsync) - importing/exporting transactions from local blockchain node storage
  * [goscan](./misc/goscan) - dumps imports from specified file’s AST
  * [genproto](./misc/genproto) - helper for generating .proto implementations
  * [gnofaucet](./gno.land/cmd/gnofaucet) - serves GNOT faucet
</details>
