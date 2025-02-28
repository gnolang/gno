# Package Paths

A package path is a unique identifier for any package that lives on the gno.land
blockchain. It consists of multiple parts separated with `/`. 
It is generally used to refer to a specific realm or pure package, through a
[call transaction](../dev-guides/gnokey/making-transactions.md#call),
an ABCI query, or a Gno import statement.

Let's take a look at a few examples of package paths:

- The gno.land home realm - [gno.land/r/gnoland/home](https://gno.land/r/gnoland/home)
- The Hall of Fame realm - [gno.land/r/leon/hof](https://gno.land/r/leon/hof)
- The AVL tree package - [gno.land/p/demo/avl](https://gno.land/p/demo/avl)

The above paths have similarities and differences. Let's break their parts down 
one by one:

- `gno.land` is the chain domain. Currently, only `gno.land` is supported, but
the ecosystem may expand in the future and other chain domains might become available.
- `p` or `r` declare the type of package found at the path. `p` stands for
[`pure` (packages)](packages.md), while `r` represents [`realm`](realms.md).
- `demo`, `gnoland`, etc., represent namespaces. Read more about Gno namespaces
[below](#gno-namespaces).
- `home`, `hof`, `avl`, etc., represent the package name found at the path. This 
part must match the package name declaration found in `.gno` files inside.

Two more important facts about package paths are the following:
- The maximum length of a package path is `256` characters.
- A realm's address is directly derived from its package path, by using 
[`std.DerivePkgAddr()`](../reference/std.md#derivepkgaddr)

## Gno Namespaces

Gno Namespaces provide users with the exclusive ability to publish code under
their designated namespaces, similar to GitHub's user and organization model.

Initially, all users are granted a default namespace with their address - a 
pseudo-anonymous (PA) namespace - to which the associated address can deploy.
This namespace has the following format:
```
gno.land/{p,r}/{std.Address}/**
```

For example, for address `g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`, all the
following paths are valid for deployments:

- `gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/mypackage` 
- `gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/myrealm`
- `gno.land/p/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/mypackage/subpackage/package` 
- `gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/subpackage/realm`
- etc.

Apart from package names, developers can define subpackages to further organize 
their code, as seen in the example above. Packages can have any varying level of 
depth as long as the full package path doesn't exceed `256` characters. 

### Registering a custom namespace

To register a custom namespace, users need to use the 
[`gno.land/r/demo/users`](https://gno.land/r/demo/users) realm.
It allows users to register a username for a fee of `200 GNOT`, which is
in turn used as a reference for the namespace of the user.

Once a username is registered, the `gno.land/r/sys/users` system mechanism enforces
that only the address associated with the registered username can deploy code 
under that namespace.

Check out "[Registering a namespace](../getting-started/interacting-with-gnoland.md)"
for a detailed guide on how to register a Gno namespace.

