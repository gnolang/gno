// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity ^0.8.27;

import {IDelegation, Signature, Delegation} from "./interfaces/IDelegation.sol";

import {ERC20PermitUpgradeable} from
    "../lib/openzeppelin-contracts-upgradeable/contracts/token/ERC20/extensions/ERC20PermitUpgradeable.sol";
import {ECDSA} from
    "../lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts/utils/cryptography/ECDSA.sol";
import {Ownable2StepUpgradeable} from
    "../lib/openzeppelin-contracts-upgradeable/contracts/access/Ownable2StepUpgradeable.sol";
import {UUPSUpgradeable} from "../lib/openzeppelin-contracts-upgradeable/contracts/proxy/utils/UUPSUpgradeable.sol";
import {ERC1967Utils} from
    "../lib/openzeppelin-contracts-upgradeable/lib/openzeppelin-contracts/contracts/proxy/ERC1967/ERC1967Utils.sol";

/// @title DelegationToken
/// @author Morpho Association
/// @custom:security-contact security@morpho.org
/// @dev Extension of ERC20 to support token delegation.
///
/// This extension keeps track of the current voting power delegated to each account. Voting power can be delegated
/// either by calling the `delegate` function directly, or by providing a signature to be used with `delegateBySig`.
///
/// This enables onchain votes on external voting smart contracts leveraging storage proofs.
///
/// By default, token balance does not account for voting power. This makes transfers cheaper. Whether an account
/// has to self-delegate to vote depends on the voting contract implementation.
abstract contract DelegationToken is IDelegation, ERC20PermitUpgradeable, Ownable2StepUpgradeable, UUPSUpgradeable {
    /* CONSTANTS */

    bytes32 internal constant DELEGATION_TYPEHASH =
        keccak256("Delegation(address delegatee,uint256 nonce,uint256 expiry)");

    // keccak256(abi.encode(uint256(keccak256("morpho.storage.DelegationToken")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 internal constant DELEGATION_TOKEN_STORAGE_LOCATION =
        0x669be2f4ee1b0b5f3858e4135f31064efe8fa923b09bf21bf538f64f2c3e1100;

    /* STORAGE LAYOUT */

    /// @custom:storage-location erc7201:morpho.storage.DelegationToken
    struct DelegationTokenStorage {
        mapping(address => address) _delegatee;
        mapping(address => uint256) _delegatedVotingPower;
        mapping(address => uint256) _delegationNonce;
    }

    /* ERRORS */

    /// @notice The signature used has expired.
    error DelegatesExpiredSignature();

    /// @notice The delegation nonce used by the signer is not its current delegation nonce.
    error InvalidDelegationNonce();

    /* EVENTS */

    /// @notice Emitted when an delegator changes their delegatee.
    event DelegateeChanged(address indexed delegator, address indexed oldDelegatee, address indexed newDelegatee);

    /// @notice Emitted when a delegatee's delegated voting power changes.
    event DelegatedVotingPowerChanged(address indexed delegatee, uint256 oldVotes, uint256 newVotes);

    /* CONSTRUCTOR */

    /// @dev Disables initializers for the implementation contract.
    constructor() {
        _disableInitializers();
    }

    /* GETTERS */

    /// @notice Returns the delegatee that `account` has chosen.
    function delegatee(address account) public view returns (address) {
        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        return $._delegatee[account];
    }

    /// @notice Returns the current voting power delegated to `account`.
    function delegatedVotingPower(address account) external view returns (uint256) {
        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        return $._delegatedVotingPower[account];
    }

    /// @notice Returns the current delegation nonce of `account`.
    function delegationNonce(address account) external view returns (uint256) {
        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        return $._delegationNonce[account];
    }

    /// @notice Returns the contract's current implementation address.
    function getImplementation() external view returns (address) {
        return ERC1967Utils.getImplementation();
    }

    /* DELEGATE */

    /// @notice Delegates the balance of the sender to `newDelegatee`.
    /// @dev Delegating to the zero address effectively removes the delegation, incidentally making transfers cheaper.
    /// @dev Delegating to the previous delegatee does not revert.
    function delegate(address newDelegatee) external {
        address delegator = _msgSender();
        _delegate(delegator, newDelegatee);
    }

    /// @notice Delegates the balance of the signer to `newDelegatee`.
    /// @dev Delegating to the zero address effectively removes the delegation, incidentally making transfers cheaper.
    /// @dev Delegating to the previous delegatee effectively revokes past signatures with the same nonce.
    function delegateWithSig(Delegation calldata delegation, Signature calldata signature) external {
        require(block.timestamp <= delegation.expiry, DelegatesExpiredSignature());

        address delegator = ECDSA.recover(
            _hashTypedDataV4(keccak256(abi.encode(DELEGATION_TYPEHASH, delegation))),
            signature.v,
            signature.r,
            signature.s
        );

        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        require(delegation.nonce == $._delegationNonce[delegator]++, InvalidDelegationNonce());

        _delegate(delegator, delegation.delegatee);
    }

    /* INTERNAL */

    /// @dev Delegates the balance of the `delegator` to `newDelegatee`.
    function _delegate(address delegator, address newDelegatee) internal {
        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        address oldDelegatee = $._delegatee[delegator];
        $._delegatee[delegator] = newDelegatee;

        emit DelegateeChanged(delegator, oldDelegatee, newDelegatee);
        _moveDelegateVotes(oldDelegatee, newDelegatee, balanceOf(delegator));
    }

    /// @dev Moves voting power when tokens are transferred.
    function _update(address from, address to, uint256 value) internal virtual override {
        super._update(from, to, value);
        _moveDelegateVotes(delegatee(from), delegatee(to), value);
    }

    /// @dev Moves delegated votes from one delegate to another.
    function _moveDelegateVotes(address from, address to, uint256 amount) internal {
        DelegationTokenStorage storage $ = _getDelegationTokenStorage();
        if (from != to && amount > 0) {
            if (from != address(0)) {
                uint256 oldValue = $._delegatedVotingPower[from];
                uint256 newValue = oldValue - amount;
                $._delegatedVotingPower[from] = newValue;
                emit DelegatedVotingPowerChanged(from, oldValue, newValue);
            }
            if (to != address(0)) {
                uint256 oldValue = $._delegatedVotingPower[to];
                uint256 newValue = oldValue + amount;
                $._delegatedVotingPower[to] = newValue;
                emit DelegatedVotingPowerChanged(to, oldValue, newValue);
            }
        }
    }

    /// @dev Returns the DelegationTokenStorage struct.
    function _getDelegationTokenStorage() internal pure returns (DelegationTokenStorage storage $) {
        assembly {
            $.slot := DELEGATION_TOKEN_STORAGE_LOCATION
        }
    }

    /// @inheritdoc UUPSUpgradeable
    function _authorizeUpgrade(address) internal override onlyOwner {}
}
