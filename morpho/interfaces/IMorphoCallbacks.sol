// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.5.0;

/// @title IMorphoLiquidateCallback
/// @notice Interface that liquidators willing to use `liquidate`'s callback must implement.
interface IMorphoLiquidateCallback {
    /// @notice Callback called when a liquidation occurs.
    /// @dev The callback is called only if data is not empty.
    /// @param repaidAssets The amount of repaid assets.
    /// @param data Arbitrary data passed to the `liquidate` function.
    function onMorphoLiquidate(uint256 repaidAssets, bytes calldata data) external;
}

/// @title IMorphoRepayCallback
/// @notice Interface that users willing to use `repay`'s callback must implement.
interface IMorphoRepayCallback {
    /// @notice Callback called when a repayment occurs.
    /// @dev The callback is called only if data is not empty.
    /// @param assets The amount of repaid assets.
    /// @param data Arbitrary data passed to the `repay` function.
    function onMorphoRepay(uint256 assets, bytes calldata data) external;
}

/// @title IMorphoSupplyCallback
/// @notice Interface that users willing to use `supply`'s callback must implement.
interface IMorphoSupplyCallback {
    /// @notice Callback called when a supply occurs.
    /// @dev The callback is called only if data is not empty.
    /// @param assets The amount of supplied assets.
    /// @param data Arbitrary data passed to the `supply` function.
    function onMorphoSupply(uint256 assets, bytes calldata data) external;
}

/// @title IMorphoSupplyCollateralCallback
/// @notice Interface that users willing to use `supplyCollateral`'s callback must implement.
interface IMorphoSupplyCollateralCallback {
    /// @notice Callback called when a supply of collateral occurs.
    /// @dev The callback is called only if data is not empty.
    /// @param assets The amount of supplied collateral.
    /// @param data Arbitrary data passed to the `supplyCollateral` function.
    function onMorphoSupplyCollateral(uint256 assets, bytes calldata data) external;
}

/// @title IMorphoFlashLoanCallback
/// @notice Interface that users willing to use `flashLoan`'s callback must implement.
interface IMorphoFlashLoanCallback {
    /// @notice Callback called when a flash loan occurs.
    /// @dev The callback is called only if data is not empty.
    /// @param assets The amount of assets that was flash loaned.
    /// @param data Arbitrary data passed to the `flashLoan` function.
    function onMorphoFlashLoan(uint256 assets, bytes calldata data) external;
}
