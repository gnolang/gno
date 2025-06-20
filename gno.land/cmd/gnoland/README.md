# `gnoland`

`gnoland` is the production binary powering the gno.land chain. You might want to run it if you're debugging low-level behavior or building advanced tooling like indexers.

For **local development**, we recommend using [gnodev](../../../contribs/gnodev) — a developer-optimized node that makes writing and testing Gno contracts much easier.

> Note: The `gnoland` binary is **specific to the gno.land chain**. Other chains in the Gno ecosystem will use different binaries tailored to their own configurations and goals.

## Getting Started

To run your own local gno.land node, follow this guide from the [gnops blog](https://gnops.io/):  
👉 [Setting up a local Gno chain from scratch](https://gnops.io/articles/guides/local-chain/)

**TODO:** Make this README self-sufficient so that we don’t depend on this blog.

### Install `gnoland`

```bash
git clone git@github.com:gnolang/gno.git
cd gno/gno.land
make install.gnoland
```

### Start a new network

```bash
gnoland start -lazy
```

### Join an existing network

- XXX: create a profile on r/gnoland/valopers
- XXX: create a validation-only session key
- XXX: get approved by govdao
- XXX: profit

### Interact

Once running, you can interact with it using:
- [gnokey](../gnokey) – CLI wallet & tool
- [gnoweb](../gnoweb) – Web-based interface
- XXX: faucet