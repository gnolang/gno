// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity 0.8.27;

import {IOptimismMintableERC20} from "./interfaces/IOptimismMintableERC20.sol";
import {IERC165} from
    "../lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts/utils/introspection/IERC165.sol";

import {DelegationToken} from "./DelegationToken.sol";

/// @title MorphoTokenOptimism
/// @author Morpho Association
/// @custom:security-contact security@morpho.org
/// @notice The Morpho token contract for Optimism networks.
contract MorphoTokenOptimism is DelegationToken, IOptimismMintableERC20 {
    /* CONSTANTS */

    /// @dev The name of the token.
    string internal constant NAME = "Morpho Token";

    /// @dev The symbol of the token.
    string internal constant SYMBOL = "MORPHO";

    /// @notice The Morpho token on Ethereum.
    /// @dev Does not follow our classic naming convention to suits Optimism' standard.
    address public immutable remoteToken;

    /// @notice The StandardBridge.
    /// @dev Does not follow our classic naming convention to suits Optimism' standard.
    address public immutable bridge;

    /* ERRORS */

    /// @notice Thrown if the address is the zero address.
    error ZeroAddress();

    /// @notice Thrown if the caller is not the bridge.
    error NotBridge();

    /* CONSTRUCTOR */

    /// @notice Construct the contract.
    /// @param newRemoteToken The remote token address.
    /// @param newBridge The bridge address.
    constructor(address newRemoteToken, address newBridge) {
        require(newRemoteToken != address(0), ZeroAddress());
        require(newBridge != address(0), ZeroAddress());

        remoteToken = newRemoteToken;
        bridge = newBridge;
    }

    /* MODIFIERS */

    /// @dev A modifier that only allows the bridge to call.
    modifier onlyBridge() {
        require(_msgSender() == bridge, NotBridge());
        _;
    }

    /* EXTERNAL */

    /// @notice Initializes the contract.
    /// @param owner The new owner.
    function initialize(address owner) external initializer {
        __ERC20_init(NAME, SYMBOL);
        __ERC20Permit_init(NAME);
        __Ownable_init(owner);
    }

    /// @dev Allows the StandardBridge on this network to mint tokens.
    function mint(address to, uint256 amount) external onlyBridge {
        _mint(to, amount);
    }

    /// @dev Allows the StandardBridge on this network to burn tokens.
    function burn(address from, uint256 amount) external onlyBridge {
        _burn(from, amount);
    }

    /// @notice ERC165 interface check function.
    /// @param _interfaceId Interface ID to check.
    /// @return Whether or not the interface is supported by this contract.
    function supportsInterface(bytes4 _interfaceId) external pure returns (bool) {
        return _interfaceId == type(IERC165).interfaceId || _interfaceId == type(IOptimismMintableERC20).interfaceId;
    }
}
