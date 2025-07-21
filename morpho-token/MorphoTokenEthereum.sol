// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity 0.8.27;

import {DelegationToken} from "./DelegationToken.sol";

/// @title MorphoTokenEthereum
/// @author Morpho Association
/// @custom:security-contact security@morpho.org
/// @notice The Morpho token contract for Ethereum.
contract MorphoTokenEthereum is DelegationToken {
    /* CONSTANTS */

    /// @dev The name of the token.
    string internal constant NAME = "Morpho Token";

    /// @dev The symbol of the token.
    string internal constant SYMBOL = "MORPHO";

    /* EXTERNAL */

    /// @notice Initializes the contract.
    /// @param owner The new owner.
    /// @param wrapper The wrapper contract address to migrate legacy MORPHO tokens to the new one.
    function initialize(address owner, address wrapper) external initializer {
        __ERC20_init(NAME, SYMBOL);
        __ERC20Permit_init(NAME);
        __Ownable_init(owner);

        _mint(wrapper, 1_000_000_000e18); // Mint 1B to the wrapper contract.
    }

    /// @notice Mints tokens.
    function mint(address to, uint256 amount) external onlyOwner {
        _mint(to, amount);
    }

    /// @notice Burns sender's tokens.
    function burn(uint256 amount) external {
        _burn(_msgSender(), amount);
    }
}
