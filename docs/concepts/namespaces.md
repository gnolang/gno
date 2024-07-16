---
id: namespaces
---

# Namespaces

Namespaces provide users with the exclusive capability to publish contracts under their designated namespaces,
similar to GitHub's user and organization model.

:::warning Not enabled

This feature isn't enabled by default on the portal loop chain and is currently available only on test4.gno.land.

:::

# Package Path

A package path is a unique identifier for each package/realm. It specifies the location of the package source
code which helps differentiate it from others. You can use a package path to:

- Call a specific function from a package/realm. (e.g using `gnokey maketx call`)
- Import it in other packages/realms.

Here's a breakdown of the structure of a package path:

- Domain: The domain of the blockchain where the package is deployed. Currently, only `gno.land/` is supported.
- Type: Defines the type of package.
    - `p/`: [Package](packages.md)
    - `r/`: [Realm](realms.md)
- Namespace: A namespace can be included after the type (e.g., user or organization name). Namespaces are a
  way to group related packages or realms, but currently ownership cannot be claimed. (see 
  [Issue#1107](https://github.com/gnolang/gno/issues/1107) for more info)
- Remaining Path: The remaining part of the path.
    - Can only contain alphanumeric characters (letters and numbers) and underscores.
    - No special characters allowed (except underscore).
    - Cannot consist solely of underscores. It must have at least one allowed alphanumeric character.
    - Cannot start with a number. It should begin with a letter.
    - Cannot end with a trailing slash (`/`).

Examples:

- `gno.land/p/demo/avl`: This signifies a package named `avl` within the `demo` namespace.
- `gno.land/r/gnoland/home`: This signifies a realm named `home` within the `gnoland` namespace.

## Registration Process

The registration process is contract-based. The `AddPkg` command references
`sys/users` for filtering, which in turn is based on `r/demo/users`.

When `sys/users` is enabled, you need to register a name using `r/demo/users`. You can call the
`r/demo/users.Register` function to register the name for the caller's address.

> ex: `test1` user registering as `patrick`
```bash
$ gnokey maketx call -pkgpath gno.land/r/demo/users \
    -func Register \
    -gas-fee 1000000ugnot -gas-wanted 2000000 \
    -broadcast \
    -chainid=test4 \
    -send=20000000ugnot \
    -args '' \
    -args 'patrick' \
    -args 'My Profile Quote' test1
```

:::note Chain-ID

Do not forget to update chain id, adequate to the network you're interacting with

:::


After successful registration, you can add a package under the registered namespace.

## Anonymous Namespace

Gno.land offers the ability to add a package without having a registered namespace. 
You can do this by using your own address as a namespace. This is formatted as `{p,r}/{std.Address}/**`. 

> ex:  with `test1` user adding a package `microblog` using his own address as namespace
```bash
$ gnokey maketx addpkg \
    --pkgpath "gno.land/r/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5/microblog" \
    --pkgdir "examples/gno.land/p/demo/microblog" \
    --deposit 100000000ugnot \
    --gas-fee 1000000ugnot \
    --gas-wanted 2000000 \
    --broadcast \
    --chainid test4 \
    test1
```
