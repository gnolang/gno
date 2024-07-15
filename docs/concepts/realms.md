---
id: realms
---

# Realms

A Realm represents an on-chain entity which has an address, can execute code
and is capable of managing Coins associated with the given address.
It boils down to one of two kinds of realms:

- User realms: an end-user of the blockchain, who has a private key from
    which the public key and address are subsequently derived.
- Code realms: a package published on-chain, with storage through globlal
    variables and access to the Banker. It is Gno's specific flavour of a
    "smart contract".

In many instances, when simply saying a "realm", the latter is intended;
however, users are also considered "realms".

Code realms offer limitless potential - they allow you to create virtually any
application while being:

- Composable: using other realms is an `import` statement away.
- Transparent: like all code on Gno.land, they must be added to the chain in
    their source format.
- Stateful: global variables are automatically persisted to the chain when
    modified in a transaction.

Some examples of practical usages of realms:

- Self-custodial financial exchanges (decentralized exchanges), like
    [Gnoswap](https://github.com/gnoswap-labs/gnoswap).
- Chess servers, like [GnoChess](https://github.com/gnolang/gnochess), and
    generally speaking game servers.
- Forum-style [boards](https://gno.land/r/demo/boards), Twitter-style
    [microblogs](https://gno.land/r/demo/microblog), and other social
    applications.
- Transparent insurance systems.
- Fair and accessible voting systems.

## Pure vs Realm Packages

#### [**Pure Packages**](https://github.com/gnolang/gno/tree/master/examples/gno.land/p)

* A unit that contains functionalities and utilities that can be used in realms.
* Packages are stateless.
* The default import path is `gno.land/p/~~~`.
* Can be imported to other realms or packages.
* Cannot import realms.

#### [**Realm Packages**](https://github.com/gnolang/gno/tree/master/examples/gno.land/r)

* Smart contracts in Gno.
* Realms are stateful.
* Realms can own assets (tokens).
* The default import path is `gno.land/r/~~~`.
* Realms can implement `Render(path string) string` to simplify dapp frontend development by allowing users to request
  markdown renderings from validators and full nodes without a transaction.

A notable feature of realms is the `Render()` function.

```go
package demo

func Render(path string) string {
	return "# Hello Gno!"
}
```

Upon calling the realm above, `# Hello Gno!` is printed with a string-typed `path` declared in an argument. It should be
noted that while the `path` argument included in the sample code is not utilized, it serves the purpose of
distinguishing the path during the rendering process.
