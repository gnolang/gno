---
id: gno-modules
---

# Gno Modules

The packages and realms containing `gno.mod` file can be referred as Gno modules. `gno.mod` file is introduced to enhance local testing and handle dependency management while testing Gno packages/realms locally. At the time of writing, `gno.mod` is only used by the `gno` tool for local development, and it is disregarded on the gno.land chain.

## What is the gno.mod file for?

`gno.mod` file is very useful for local testing and development. Its primary purposes include:

- **Working outside of the monorepo**: by adding a `gno.mod` file to your directory, all gno tooling will recognise it and understand the implicit import path of your current directory (marked by the `module` directive in your `gno.mod` file).
- **Local dependency management**: the gno.mod file allows you to manage and download local dependencies effectively when developing Go Modules.
- **Configuration and metadata (WIP)**: while the gno.mod file is currently used for specifying dependencies, it's worth noting that in the future, it might also serve as a container for additional configuration and metadata related to Gno Modules. For more information, see: [issue #498](https://github.com/gnolang/gno/issues/498).

## Gno Modules and Subdirectories

It's important to note that Gno Modules do not include subdirectories. Each directory within your project is treated as an individual Gno Module, and each should contain its own gno.mod file, even if it's located within an existing Gno Module directory.

## Available gno Commands

The gno command-line tool provides several commands to work with the gno.mod file and manage dependencies in Gno Modules:

- **gno mod init**: small helper to initialize a new `gno.mod` file.
- **gno mod download**: downloads the dependencies specified in the gno.mod file. This command fetches the required dependencies from chain and ensures they are available for local testing and development.
- **gno mod why**: explains why the specified package or module is being kept by `gno mod tidy`.

## Sample `gno.mod` file

```
module gno.land/p/demo/sample

```

- **`module gno.land/p/demo/sample`**: specifies the package/realm import path.

