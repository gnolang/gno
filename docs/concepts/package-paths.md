---
id: package-paths
---

# Package Paths

A package path is a unique identifier for each package/realm. It specifies the 
location of the package source code which helps differentiate it from others. 
You can use a package path to:

- Call a specific function from a package/realm. (e.g using `gnokey maketx call`)
- Import it in other packages/realms.

Here's a breakdown of the structure of a package path:

- Domain: The domain of the blockchain where the package is deployed. 
Currently, only `gno.land/` is supported.
- Type: Defines the type of package.
    - `p/`: [Package](packages.md)
    - `r/`: [Realm](realms.md)
- Namespace: A namespace can be included after the type (e.g., user or organization name). Namespaces are a
way to group related packages or realms. A user can register a namespace for 
themselves, or use the anonymous namespace. Read more [below](#namespaces).

Here are some examples of package paths:

- `gno.land/p/demo/avl`: This signifies a package named `avl` within the `demo` namespace.
- `gno.land/r/gnoland/home`: This signifies a realm named `home` within the `gnoland` namespace.

## Namespaces

Namespaces provide users with the exclusive ability to publish code under
their designated namespaces, similar to GitHub's user and organization model.
A namespace follows these rules:
  - Needs to be between 6 and 17 characters.
  - Can only contain alphanumeric characters.
  - No special characters are allowed (except underscore).
  - Cannot consist solely of underscores. A namespace must have at least one 
allowed alphanumeric character.
  - Cannot start with a number. A namespace must begin with a letter.
  - Cannot end with a trailing slash (`/`).

:::warning Namespaces on gno.land testnets

This feature is currently only enabled on the [Test4 testnet](./testnets.md#test4).
Other networks, such as the Portal Loop, do not have this feature enabled.

:::

## Registering a namespace

To register a namespace, you need to use the `r/demo/users` realm. It allows
users to register a username for a fee of 200 `GNOTs`, which is in turn used as 
a reference for the namespace of the user. 

Once a username is registered, `r/sys/users` is used as a filtering mechanism 
which will allow code deployments from the registering address to the namespace
matching the username. 

A username can be registered by calling the `Register()` function in `r/demo/users`.
The `Register()` function will also allow you to add a string as a description 
of your profile. Check out [`r/demo/users` on Test4](https://test4.gno.land/r/demo/users).

For example, the following `gnokey` command will register `patrick` as the username
for the address `mykey`.

```bash
$ gnokey maketx call -pkgpath gno.land/r/demo/users \
    -func Register \
    -gas-fee 1000000ugnot -gas-wanted 2000000 \
    -broadcast \
    -chainid=test4 \
    -remote https://rpc.test4.gno.land
    -send=20000000ugnot \
    -args '' \
    -args 'patrick' \
    -args 'My Profile Quote' mykey
```

:::note Interacting with the correct network

Make sure to use the proper chain ID and remote when interacting with a 
gno.land network. Check out the [Network Configurations](../reference/network-config.md)
page for a list of all available endpoints.

:::

After successful registration, you can add a package under your registered namespace.

## Anonymous Namespace

gno.land offers the ability to deploy code without having a registered namespace. 
You can do this by using your own address as a namespace. This is formatted as `{p,r}/{std.Address}/**`. 

For example, with `mykey` being `g1zmgrq5y2vxuqc8shkuc0vr5dj23eklf2xr720x`, 
the following deployment would work:

```bash
$ gnokey maketx addpkg \
    --pkgpath "gno.land/r/g1zmgrq5y2vxuqc8shkuc0vr5dj23eklf2xr720x/myrealm" \
    --pkgdir "." \
    --deposit 100000000ugnot \
    --gas-fee 1000000ugnot \
    --gas-wanted 2000000 \
    --broadcast \
    --chainid test4 \
    -remote https://rpc.test4.gno.land
    mykey
```
