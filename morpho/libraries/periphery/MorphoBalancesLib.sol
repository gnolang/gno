// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity ^0.8.0;

import {Id, MarketParams, Market, IMorpho} from "../../interfaces/IMorpho.sol";
import {IIrm} from "../../interfaces/IIrm.sol";

import {MathLib} from "../MathLib.sol";
import {UtilsLib} from "../UtilsLib.sol";
import {MorphoLib} from "./MorphoLib.sol";
import {SharesMathLib} from "../SharesMathLib.sol";
import {MarketParamsLib} from "../MarketParamsLib.sol";

/// @title MorphoBalancesLib
/// @author Morpho Labs
/// @custom:contact security@morpho.org
/// @notice Helper library exposing getters with the expected value after interest accrual.
/// @dev This library is not used in Morpho itself and is intended to be used by integrators.
/// @dev The getter to retrieve the expected total borrow shares is not exposed because interest accrual does not apply
/// to it. The value can be queried directly on Morpho using `totalBorrowShares`.
library MorphoBalancesLib {
    using MathLib for uint256;
    using MathLib for uint128;
    using UtilsLib for uint256;
    using MorphoLib for IMorpho;
    using SharesMathLib for uint256;
    using MarketParamsLib for MarketParams;

    /// @notice Returns the expected market balances of a market after having accrued interest.
    /// @return The expected total supply assets.
    /// @return The expected total supply shares.
    /// @return The expected total borrow assets.
    /// @return The expected total borrow shares.
    function expectedMarketBalances(IMorpho morpho, MarketParams memory marketParams)
        internal
        view
        returns (uint256, uint256, uint256, uint256)
    {
        Id id = marketParams.id();
        Market memory market = morpho.market(id);

        uint256 elapsed = block.timestamp - market.lastUpdate;

        // Skipped if elapsed == 0 or totalBorrowAssets == 0 because interest would be null, or if irm == address(0).
        if (elapsed != 0 && market.totalBorrowAssets != 0 && marketParams.irm != address(0)) {
            uint256 borrowRate = IIrm(marketParams.irm).borrowRateView(marketParams, market);
            uint256 interest = market.totalBorrowAssets.wMulDown(borrowRate.wTaylorCompounded(elapsed));
            market.totalBorrowAssets += interest.toUint128();
            market.totalSupplyAssets += interest.toUint128();

            if (market.fee != 0) {
                uint256 feeAmount = interest.wMulDown(market.fee);
                // The fee amount is subtracted from the total supply in this calculation to compensate for the fact
                // that total supply is already updated.
                uint256 feeShares =
                    feeAmount.toSharesDown(market.totalSupplyAssets - feeAmount, market.totalSupplyShares);
                market.totalSupplyShares += feeShares.toUint128();
            }
        }

        return (market.totalSupplyAssets, market.totalSupplyShares, market.totalBorrowAssets, market.totalBorrowShares);
    }

    /// @notice Returns the expected total supply assets of a market after having accrued interest.
    function expectedTotalSupplyAssets(IMorpho morpho, MarketParams memory marketParams)
        internal
        view
        returns (uint256 totalSupplyAssets)
    {
        (totalSupplyAssets,,,) = expectedMarketBalances(morpho, marketParams);
    }

    /// @notice Returns the expected total borrow assets of a market after having accrued interest.
    function expectedTotalBorrowAssets(IMorpho morpho, MarketParams memory marketParams)
        internal
        view
        returns (uint256 totalBorrowAssets)
    {
        (,, totalBorrowAssets,) = expectedMarketBalances(morpho, marketParams);
    }

    /// @notice Returns the expected total supply shares of a market after having accrued interest.
    function expectedTotalSupplyShares(IMorpho morpho, MarketParams memory marketParams)
        internal
        view
        returns (uint256 totalSupplyShares)
    {
        (, totalSupplyShares,,) = expectedMarketBalances(morpho, marketParams);
    }

    /// @notice Returns the expected supply assets balance of `user` on a market after having accrued interest.
    /// @dev Warning: Wrong for `feeRecipient` because their supply shares increase is not taken into account.
    /// @dev Warning: Withdrawing using the expected supply assets can lead to a revert due to conversion roundings from
    /// assets to shares.
    function expectedSupplyAssets(IMorpho morpho, MarketParams memory marketParams, address user)
        internal
        view
        returns (uint256)
    {
        Id id = marketParams.id();
        uint256 supplyShares = morpho.supplyShares(id, user);
        (uint256 totalSupplyAssets, uint256 totalSupplyShares,,) = expectedMarketBalances(morpho, marketParams);

        return supplyShares.toAssetsDown(totalSupplyAssets, totalSupplyShares);
    }

    /// @notice Returns the expected borrow assets balance of `user` on a market after having accrued interest.
    /// @dev Warning: The expected balance is rounded up, so it may be greater than the market's expected total borrow
    /// assets.
    function expectedBorrowAssets(IMorpho morpho, MarketParams memory marketParams, address user)
        internal
        view
        returns (uint256)
    {
        Id id = marketParams.id();
        uint256 borrowShares = morpho.borrowShares(id, user);
        (,, uint256 totalBorrowAssets, uint256 totalBorrowShares) = expectedMarketBalances(morpho, marketParams);

        return borrowShares.toAssetsUp(totalBorrowAssets, totalBorrowShares);
    }
}
