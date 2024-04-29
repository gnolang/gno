---
id: namespaces
---

# Namespaces

Namespaces provide users with the exclusive capability to publish contracts under their designated namespaces, similar to GitHub's user and organization model.

This feature is currently a work in progress (WIP). To learn more about namespaces, please checkout https://github.com/gnolang/gno/issues/1107.

# Package Path

A package path is a unique identifier for each package/realm. It specifies the location of the package source code which helps differentiate it from others. You can use a package path to:

- Call a specific function from a package/realm. (e.g using `gnokey maketx call`)
- Import it in other packages/realms.

Here's a breakdown of the structure of a package path:

- Domain: The domain of the blockchain where the package is deployed. Currently, only `gno.land/` is supported.
- Type: Defines the type of package.
    - `p/`: [Package](packages.md)
    - `r/`: [Realm](realms.md)
- Namespace: A namespace can be included after the type (e.g., user or organization name). Namespaces are a way to group related packages or realms, but currently ownership cannot be claimed. (see [Issue #1107](https://github.com/gnolang/gno/issues/1107) for more info)
- Remaining Path: The remaining part of the path.
    - Can only contain alphanumeric characters (letters and numbers) and underscores.
    - No special characters allowed (except underscore).
    - Cannot consist solely of underscores. It must have at least one allowed alphanumeric character.
    - Cannot start with a number. It should begin with a letter.
    - Cannot end with a trailing slash (`/`).

Examples:

- `gno.land/p/demo/avl`: This signifies a package named `avl` within the `demo` namespace.
- `gno.land/r/gnoland/home`: This signifies a realm named `home` within the `gnoland` namespace.
