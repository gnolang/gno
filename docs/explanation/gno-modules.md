---
id: gno-modules
---

# Gno Modules

The packages and realms containing `gno.mod` file can be reffered as Gno modules. `gno.mod` file is introduced to enhance local testing and handle dependency management while testing Gno packages/realms locally. While it primarily serves local development and testing purposes, blockchain do not recognize or utilize this file directly.

## What Is the gno.mod File for?

`gno.mod` file is a critical for local testing and development. Its primary purposes include:

- **Local Dependency Management**: The gno.mod file allows you to manage local dependencies effectively when developing God Modules. This facilitates testing and iterative development.

- **Module Sorting while Publishing**: Additionally, the gno.mod file is used to sort and automatically publish modules that are located within the `/examples` directory to the blockchain when the chain starts.

- **Configuration and Metadata (Potential Future Use)**: While the gno.mod file is currently used for specifying dependencies, it's worth noting that in the future, it might also serve as a container for additional configuration and metadata related to Gno Modules. This could include information such as module descriptions, veriosn, authorship details, or licensing information. (See: https://github.com/gnolang/gno/issues/498)

## Gno Modules and Subdirectories

It's important to note that Gno Modules do not include subdirectories. Each directory within your project is treated as an individual Gno Module, and each should contain its own gno.mod file, even if it's located within an existing Gno Module directory.

## Available gno Commands

The gno command-line tool provides several commands to work with the gno.mod file and manage dependencies in Gno Modules:

- `gno mod init`: Initializes a new `gno.mod` file. Allowing you to specify dependencies.

- **gno mod download**: Downloads the dependencies specified in the gno.mod file. This command fetches the required dependencies from chain and ensures they are available for local testing and development.

- **gno mod tidy**: This command helps maintain the cleanliness of the gno.mod file by removing any unused dependencies and automatically adds any dependencies that are required but not yet listed in the gno.mod file. It ensures that your gno.mod file remains up-to-date and free from unnecessary clutter, eliminating the need for manual maintenance.

## Sample `gno.mod` file

```
module gno.land/p/demo/sample

require (
    gno.land/p/demo/avl v0.0.0-latest
    gno.land/p/demo/testutils v0.0.0-latest
)

```

- **`module gno.land/p/demo/sample`**: Specifies the package/realm import path.

- **`require` Block**: Lists the required dependencies. Here using the latest available versions of "gno.land/p/demo/avl" and "gno.land/p/demo/testutils". These dependencies should be specified with the version "v0.0.0-latest" since as of now on-chain packages do not support versioning.
