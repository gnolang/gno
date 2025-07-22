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
// Use this package to manage the Volos DAO. Membership updates are automatic
// via staking actions. See commondao.gno for core DAO logic, vls.gno/xvls.gno
// for token contracts, and staker.gno for staking logic.
package governance
