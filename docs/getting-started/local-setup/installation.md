---
id: installation
---

# Installation

## Overview
In this tutorial, you will learn how to set up the Gno development environment 
locally, so you can get up and running writing Gno code. You will download and 
install all the necessary tooling, and validate that it is correctly configured
to run on your machine.

## Prerequisites
- **Git**
- **`make` (for running Makefiles)**
- **Go 1.19+**
- **Go Environment Setup**:
  - Make sure `$GOPATH` is well-defined, and `$GOPATH/bin` is added to your `$PATH` variable.
  - To do this, you can add the following line to your `.bashrc`, `.zshrc` or other config file:
```
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH
```

## 1. Cloning the repository
To get started with a local Gno.land development environment, you must clone the
GitHub repository somewhere on disk:

```bash
git clone https://github.com/gnolang/gno.git
```

## 2. Installing the `gno` development toolkit
Next, we are going to build and install the `gno` development toolkit.
`gno` provides ample functionality to the user, among which is running,
transpiling, testing and building `.gno` files.

To install the toolkit, navigate to the `gnovm` folder from the repository root,
and run the `install` make directive:

```bash
cd gnovm
make install
```

To verify the `gno` binary is installed system-wide, you can run:

```bash
gno --help
```

You should get the help output from the command:

![gno help](../../assets/getting-started/local-setup/local-setup/gno-help.gif)

Alternatively, if you don't want to have the binary callable system-wide, you can run the binary directly:

```bash
go run ./cmd/gno --help
```

## 3. Installing other `gno` tools
The next step is to install two other tools that are required for the Gno 
development environment:

- `gnodev` - the Gno [development helper](../../gno-tooling/cli/gnodev.md)
- `gnokey` - the Gno [private key manager](working-with-key-pairs.md)

To build these tools, navigate to the root folder, and run the following:

```bash
make install.gnodev install.gnokey
```

To verify that the `gnodev` binary is installed system-wide, you can run:

```bash
gnodev
```

You should get the following output:
![gnodev](../../assets/getting-started/local-setup/local-setup/gnodev.gif)

Finally, to verify that the `gnokey` binary is installed system-wide, you can run:

```bash
gnokey --help
```

You should get the help output from the command:

![gnokey help](../../assets/getting-started/local-setup/local-setup/gnokey-help.gif)

## Conclusion

That's it 🎉

You have successfully built out and installed the necessary tools for Gno development!

In further documents, you will gain a better understanding on how they are used to make Gno work.
