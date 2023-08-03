# Gnolang examples

Folder contains Gnolang realms and libraries demos.
Share contracts here to improve engine testing, although it's not required.
Consider separate repository for contracts, but this may limit experience due to ongoing gnomod support work.
Main repository can't reference separate code, causing potential development issues.

## Usage

Our recommendation is to use the [gno](../gnovm/cmd/gno) utility to develop contracts locally before publishing them on-chain.
This approach offers a faster and streamlined workflow, along with additional debugging features.
Simply fork or create new contracts and refer to the Makefile.
Once everything looks good locally, you can then publish it on a localnet or testnet.

See [`awesome-gno` tutorials](https://github.com/gnolang/awesome-gno#tutorials).
