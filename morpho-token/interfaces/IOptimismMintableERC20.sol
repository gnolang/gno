// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.5.0;

import {IERC165} from
    "../../lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts/utils/introspection/IERC165.sol";

/// @title IOptimismMintableERC20
/// @author Morpho Association
/// @custom:security-contact security@morpho.org
/// @notice This interface is available on the OptimismMintableERC20 contract.
///         We declare it as a separate interface so that it can be used in
///         custom implementations of OptimismMintableERC20.
interface IOptimismMintableERC20 is IERC165 {
    function remoteToken() external view returns (address);

    function bridge() external returns (address);

    function mint(address to, uint256 amount) external;

    function burn(address from, uint256 amount) external;
}
