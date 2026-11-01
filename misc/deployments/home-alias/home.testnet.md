# Onyx, the Gno.land testnet

Onyx runs mainnet's code with free test tokens. Try your packages here before you deploy them to [mainnet](https://gno.land).

> [!NOTE]
> Testnet GNOT has no value, and package names here are not reserved for production. State carries over from one upgrade to the next.

---

<gno-columns>

## <gno-icon name="code-outline" /> Deploy a realm

From zero to a live realm.

<gno-button href="https://docs.gno.land/builders/getting-started" label="Read the guide" />

1. [Install gnokey](https://docs.gno.land/builders/install) and add a key
2. [Get test GNOT](https://faucet.gno.land)
3. Deploy with [`addpkg`](#deploy-from-your-terminal)
4. Call it in its Actions tab

<gno-columns-sep>

## <gno-icon name="browser" /> Test an app

Try apps with free tokens.

<gno-button href="https://faucet.gno.land" label="Open the faucet" />

1. [Install Adena](https://www.adena.app/)
2. Switch it to Onyx
3. Get test GNOT
4. Pick a realm from [the list](/r/)

<gno-columns-sep>

## <gno-icon name="lightbulb" /> Build with AI

Let your agent deploy Gno.

<gno-button href="https://mcp.gno.dev" label="Set up gnomcp" />

1. Install gnomcp
2. It creates its own key
3. It funds itself
4. It deploys with its key

</gno-columns>

---

<!-- whats-new:start -->
<!-- whats-new:end -->

## <gno-icon name="rocket" /> Deploy from your terminal

```sh
gnokey add alice   # gnokey list prints your g1 address
# Fund that address at https://faucet.gno.land (pick Onyx)

cd myrealm         # your .gno files, in package myrealm
gno mod init gno.land/r/<your-g1-addr>/myrealm

gnokey maketx addpkg \
  -pkgpath "gno.land/r/<your-g1-addr>/myrealm" \
  -pkgdir . \
  -gas-fee 40000ugnot -gas-wanted 20000000 \
  -chainid onyx-1 -remote https://rpc.onyx.testnets.gno.land:443 \
  alice
```

A new package is enabled a few seconds after `addpkg`. If it is not, `https://gpao.onyx.testnets.gno.land/status/gno.land/r/<your-g1-addr>/myrealm` says why. Still `unknown` after a minute? [Ask on Discord](https://discord.com/invite/gnoland).

---

<gno-columns>

## <gno-icon name="book" /> Learn

- [Write Gno in the browser](https://play.gno.land)
- [The Gno language](/gnolang)
- [Local development (gnodev)](https://docs.gno.land/resources/gnodev)
- [Networks documentation](https://docs.gno.land/resources/gnoland-networks/)

<gno-columns-sep>

## <gno-icon name="search-outline" /> Read the code

- [Official realms (r/gnoland)](/r/gnoland)
- [System realms (r/sys)](/r/sys)
- [Standard packages (p/nt)](/p/nt)
- [GovDAO](/r/gov/dao)

<gno-columns-sep>

## <gno-icon name="users" /> Community

- [Ask on Discord](https://discord.com/invite/gnoland)
- [Awesome Gno projects](https://github.com/gnoverse/awesome-gno)
- [Contribute](/contribute)
- [Get a grant](https://github.com/gnolang/grants)

</gno-columns>
