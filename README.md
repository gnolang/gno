# Gno

At first, there was Bitcoin, out of entropy soup of the greater All.
Then, there was Ethereum, which was created in the likeness of Bitcoin,
but made Turing complete.

Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
Then came Gno upon Cosmos and there spring forth Gnoland,
simulated by the Gnomes of the Greater Resistance.

<b>This README is a placeholder, check back again for updates</b>

## Language Features

 * Like interpreted Go, but more ambitious.
 * Completely deterministic, for complete accountability.
 * Transactional persistence across data realms.
 * Designed for concurrent blockchain smart contracts systems.
 
## Status

_Update Aug 26th, 2021: SDK/store,baseapp ported; Plan updated_

Cosmos-SDK's store and baseapp modules have been ported.
Now porting x/auth, for minimal auth usage.
Plan updated with premine distribution for GNO adoption.

_Update Aug 16th, 2021: basic file tests pass_

Basic Go file tests now pass.  Working on realm/ownership logic under tests/files/zrealm\*.go.

_Update Jul 22nd, 2021: create pkgs/crypto/keys/client as crypto wallet._

The new wallet will be used for signed communications.

_Update Jul ?, 2021: Public invited to contribute to Gnolang/files tests.

_Update Feb 13th, 2021: Implemented Logos UI framework._

This is a still a work in a progress, though much of the structure of the interpreter
and AST have taken place.  Work is ongoing now to demonstrate the Realm concept before
continuing to make the tests/files/\*.go tests pass.

Make sure you have >=[go1.15](https://golang.org/doc/install) installed, and then try this: 

```bash
> git clone git@github.com:gnolang/gno.git
> cd gno
> make test
```

## Resources

 * [GnoKey Client Tool](/cmd/gnokey) universal Gno client
 * [Amino](/pkgs/amino) complete with .proto generation
 * [BFT Consensus](/pkgs/bft) minimal port of Tendermint
 * [SDK](/pkgs/sdk) minimal port of Cosmos-SDK
 * [Logos Browser](/logos) future terminal browser
 * [Plan](/PLAN.md) project plan
 * [Roadmap](/ROADMAP.md) development roadmap
 * [Philosophy](/PHILOSOPHY.md) project philosophy

## Contact

If you can read this, the project is evolving (fast) every day.  Check
"github.com/gnolang/gno" and @jaekwon frequently.

The best way to reach out right now is to create an issue on github, but this
will change soon.
