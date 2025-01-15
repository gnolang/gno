# MultiSig Smart Contract Wallet

This package will provide functionality where multiple owners will need to approve a transaction before it is executed. It will support managing both coins and GRC20 tokens.

## Features
- **Multisig Approval**: Multiple owners must approve a transaction before it is executed.
- **Owner Verification**: Only owners can perform specific actions, which will be determined in the future.
- **Coin and GRC20 Support**: The package will manage both regular coins and GRC20 tokens.
- **Multisig Calling of Functions**: There will be functions that restrict certain actions in the contract, allowing them to be called only if they receive approval from all the owners. This is similar to the `AssertCallerIsOwner` function from the Ownable package, but it will check if all owners have approved.
