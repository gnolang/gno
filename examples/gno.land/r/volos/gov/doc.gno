// Package gov provides the entry point and coordination for the Volos protocol's
// on-chain governance system.
//
// The Volos governance system is composed of several tightly integrated
// contracts. vls.gno implements the VLS governance token, which is
// freely transferable and used for staking. The staker.gno contract manages
// staking and delegation: users stake VLS to mint xVLS to themselves or a
// delegatee, and unstaking triggers a cooldown before VLS can be withdrawn.
// Only the staker contract can mint or burn xVLS.
//
// xvls.gno defines the xVLS token, a non-transferable representation of staked
// VLS. xVLS determines voting power in governance. The governance contract
// contains the DAO logic, built on commondao.gno, where only xVLS holders are
// members and can propose or vote. Membership is updated automatically by
// staking actions.
//
// Together, these contracts enable secure, weighted, and flexible on-chain
// governance for the Volos protocol, with clear separation of concerns between
// token logic, staking and delegation, and proposal and voting mechanisms.
package gov 
