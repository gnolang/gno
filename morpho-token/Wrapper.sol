// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity 0.8.27;

import {IERC20} from
    "../lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts/token/ERC20/IERC20.sol";

/// @title Wrapper
/// @author Morpho Association
/// @custom:security-contact security@morpho.org
/// @notice The Wrapper contract to migrate from legacy MORPHO tokens.
contract Wrapper {
    /* CONSTANTS */

    /// @notice The address of the legacy Morpho token.
    address public constant LEGACY_MORPHO = 0x9994E35Db50125E0DF82e4c2dde62496CE330999;

    /* IMMUTABLES */

    /// @notice The address of the new Morpho token.
    address public immutable NEW_MORPHO;

    /* ERRORS */

    /// @notice Reverts if the address is the zero address.
    error ZeroAddress();

    /// @notice Reverts if the address is the contract address.
    error SelfAddress();

    /* CONSTRUCTOR */

    /// @dev morphoToken address can be precomputed using create2.
    constructor(address morphoToken) {
        require(morphoToken != address(0), ZeroAddress());

        NEW_MORPHO = morphoToken;
    }

    /* EXTERNAL */

    /// @dev Compliant to `ERC20Wrapper` contract from OZ for convenience.
    function depositFor(address account, uint256 value) external returns (bool) {
        require(account != address(0), ZeroAddress());
        require(account != address(this), SelfAddress());

        IERC20(LEGACY_MORPHO).transferFrom(msg.sender, address(this), value);
        IERC20(NEW_MORPHO).transfer(account, value);
        return true;
    }

    /// @dev Compliant to `ERC20Wrapper` contract from OZ for convenience.
    function withdrawTo(address account, uint256 value) external returns (bool) {
        require(account != address(0), ZeroAddress());
        require(account != address(this), SelfAddress());

        IERC20(NEW_MORPHO).transferFrom(msg.sender, address(this), value);
        IERC20(LEGACY_MORPHO).transfer(account, value);
        return true;
    }

    /// @dev To ease wrapping via the bundler contract:
    /// https://github.com/morpho-org/morpho-blue-bundlers/blob/main/src/ERC20WrapperBundler.sol
    function underlying() external pure returns (address) {
        return LEGACY_MORPHO;
    }
}
