// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity ^0.8.0;

import {IMorpho, Id} from "../../interfaces/IMorpho.sol";
import {MorphoStorageLib} from "./MorphoStorageLib.sol";

/// @title MorphoLib
/// @author Morpho Labs
/// @custom:contact security@morpho.org
/// @notice Helper library to access Morpho storage variables.
/// @dev Warning: Supply and borrow getters may return outdated values that do not include accrued interest.
library MorphoLib {
    function supplyShares(IMorpho morpho, Id id, address user) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.positionSupplySharesSlot(id, user));
        return uint256(morpho.extSloads(slot)[0]);
    }

    function borrowShares(IMorpho morpho, Id id, address user) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.positionBorrowSharesAndCollateralSlot(id, user));
        return uint128(uint256(morpho.extSloads(slot)[0]));
    }

    function collateral(IMorpho morpho, Id id, address user) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.positionBorrowSharesAndCollateralSlot(id, user));
        return uint256(morpho.extSloads(slot)[0] >> 128);
    }

    function totalSupplyAssets(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketTotalSupplyAssetsAndSharesSlot(id));
        return uint128(uint256(morpho.extSloads(slot)[0]));
    }

    function totalSupplyShares(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketTotalSupplyAssetsAndSharesSlot(id));
        return uint256(morpho.extSloads(slot)[0] >> 128);
    }

    function totalBorrowAssets(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketTotalBorrowAssetsAndSharesSlot(id));
        return uint128(uint256(morpho.extSloads(slot)[0]));
    }

    function totalBorrowShares(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketTotalBorrowAssetsAndSharesSlot(id));
        return uint256(morpho.extSloads(slot)[0] >> 128);
    }

    function lastUpdate(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketLastUpdateAndFeeSlot(id));
        return uint128(uint256(morpho.extSloads(slot)[0]));
    }

    function fee(IMorpho morpho, Id id) internal view returns (uint256) {
        bytes32[] memory slot = _array(MorphoStorageLib.marketLastUpdateAndFeeSlot(id));
        return uint256(morpho.extSloads(slot)[0] >> 128);
    }

    function _array(bytes32 x) private pure returns (bytes32[] memory) {
        bytes32[] memory res = new bytes32[](1);
        res[0] = x;
        return res;
    }
}
