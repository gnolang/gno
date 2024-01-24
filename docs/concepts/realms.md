---
id: realms
---

# Realms

A realm is a smart contract written in [Gno](./gno-language.md). The most important characteristics of realms are the following:
- Realms are stateful,
- Realms can own assets ([coins](./standard-library/coin.md)),
- Each realm is deployed under a unique package path, i.e. `gno.land/r/blog`, and also has a Gno
address derived from it, i.e. `g1n2j0gdyv45aem9p0qsfk5d2gqjupv5z536na3d`,
- Realms are deployed with a package path beginning with `gno.land/r/`,
- Realms can import packages from `gno.land/p/` to gain more functionality,

The potentials of realms are endless - you can create virtually any
application in your mind with built-in composability,
transparency, and censorship resistance. Here are some ideas of what you can build with realms:
- Complex DeFi applications
- Censorship-resistant social networks
- Fair and accessible voting systems
- Logistics and supply chain networks

You can find illustrative examples of realms either in the Gno.land monorepo, 
located within the [examples folder](https://github.com/gnolang/gno/tree/master/examples), or on-chain, under the `gno.land/r/` directory.

## Realms in code

Realms are represented by a Realm type in Gno:
```go
type Realm struct {
    addr    Address
    pkgPath string
}
```
For the full Realm API, see the [reference page](../reference/standard-library/std/realm.md).





