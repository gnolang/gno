// Package governance implements the Volos protocol's on-chain governance
// system, built atop commondao.gno. It manages membership, proposals, and
// voting, integrating tightly with the protocol's tokenomics.
//
// Membership is determined by holding xVLS, a non-transferable voting token
// minted by staking VLS via the staker contract. Staking adds members;
// fully unstaking removes them. Only xVLS holders can vote or propose.
//
// Proposal and voting logic is provided by commondao.gno, with voting power
// proportional to xVLS held at voting time. This package extends commondao
// with Volos-specific membership and token integration. Staking and delegation
// are handled externally (see staker.gno).
//
// Quorum Requirements:
// - Voting Power Quorum: Minimum total xVLS voting power required for a proposal to pass
// The voting power quorum requirement must be met for a proposal to pass, in addition to having
// more YES votes than NO votes.
//
// Quorum Disabling:
// - Setting voting power quorum to 0 or a negative value disables the quorum requirement
// - When disabled, the quorum check is skipped during proposal tallying
// - This allows for flexible governance configurations where no quorum requirements are needed
//
// Quorum Changes:
// - Quorum changes only apply to proposals created after the change
// - Active proposals continue to use the quorum value that existed when they were created
// - This ensures fairness and prevents strategic manipulation of active proposals
//
// Use this package to manage the Volos DAO. Membership updates are automatic
// via staking actions. See commondao.gno for core DAO logic, vls.gno/xvls.gno
// for token contracts, and staker.gno for staking logic.
package governance
