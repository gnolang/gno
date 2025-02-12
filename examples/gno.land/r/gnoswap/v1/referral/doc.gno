// Package referral implements a referral system on Gno. It allows
// any authorized caller to register, update, or remove referral
// information. A referral link is defined as a mapping from one
// address (the "user") to another address (the "referrer").
//
// ## Overview
//
// The referral package is composed of the following components:
//
//  1. **errors.gno**: Provides custom error types (ReferralError) with
//     specific error codes and messages.
//  2. **utils.gno**: Contains utility functions for permission checks,
//     especially isValidCaller, which ensures only specific, pre-authorized
//     callers (e.g., governance or router addresses) can invoke the core
//     functions.
//  3. **types.gno**: Defines core constants for event types, attributes,
//     and the ReferralKeeper interface, which outlines the fundamental
//     methods of the referral system (Register, Update, Remove, etc.).
//  4. **keeper.gno**: Implements the actual business logic behind the
//     ReferralKeeper interface. It uses an AVL Tree (avl.Tree) to store
//     referral data (address -> referrer). The keeper methods emit events
//     when a new referral is registered, updated, or removed.
//  5. **referral.gno**: Exposes a public API (the Referral struct)
//     that delegates to the keeper, providing external contracts or
//     applications a straightforward way to interact with the system.
//
// ## Workflow
//
// Typical usage of this contract follows these steps:
//
//  1. A caller with valid permissions invokes Register, Update, or Remove
//     through the Referral struct.
//  2. The Referral struct forwards the request to the internal keeper
//     methods.
//  3. The keeper checks caller permission (via isValidCaller), validates
//     addresses, and stores or removes data in the AVL Tree.
//  4. An event is emitted for off-chain or cross-module notifications.
//
// ## Integration with Other Contracts
//
// Other contracts can leverage the referral system in two major ways:
//
//  1. **Direct Calls**: If you wish to directly call this contract,
//     instantiate the Referral object (via NewReferral) and invoke its
//     methods, assuming you meet the authorized-caller criteria.
//
//  2. **Embedded or Extended**: If you have a complex module that includes
//     referral features, import this package and embed a Referral instance
//     in your own keeper. This way, you can handle additional validations
//     or custom logic before delegating to the existing referral functions.
//
// ## Error Handling
//
// The package defines several error types through ReferralError:
//   - `ErrInvalidAddress`: Returned when an address format is invalid
//   - `ErrUnauthorized`: Returned when the caller lacks permission
//   - `ErrNotFound`: Returned when attempting to get a non-existent referral
//   - `ErrZeroAddress`: Returned when attempting operations with zero address
//
// ## Example: Integration with a Staking Contract
//
// Suppose you have a staking contract that wants to reward referrers
// when a new user stakes tokens:
//
// ```go
//
//	import (
//	    "std"
//	    "gno.land/r/gnoswap/v1/referral"
//	    "gno.land/p/demo/mystaking"        // example staking contract
//	)
//
//	func rewardReferrerOnStake(user std.Address, amount int) {
//	    // 1) Access the referral system
//	    r := referral.NewReferral()
//
//	    // 2) Get the user's referrer
//	    refAddr, err := r.GetReferral(user)
//	    if err != nil {
//	        // handle error or skip if not found
//	        return
//	    }
//
//	    // 3) Reward the referrer
//	    mystaking.AddReward(refAddr, calculateReward(amount))
//	}
//
// ```
//
// In this simple example, the staking contract checks if the user has
// a referrer by calling `GetReferral`. If a referrer is found, it then
// calculates a reward based on the staked amount.
//
// ## Limitations and Constraints
//
// - A user can have only one referrer at a time
// - Once a referral is removed, it cannot be automatically restored
// - Only authorized contracts can modify referral relationships
// - Address validation is strict and requires proper Bech32 format
//
// # Notes
//
//   - The contract strictly enforces caller restrictions via isValidCaller.
//     Make sure to configure it to permit only the addresses or roles that
//     should be able to register or update referrals.
//   - Zero addresses are treated as a trigger for removing a referral record.
//   - The system emits events (register_referral, update_referral, remove_referral)
//     which can be consumed by other on-chain or off-chain services.
package referral
