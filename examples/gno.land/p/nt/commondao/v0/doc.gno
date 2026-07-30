// v0 - Unaudited: This is an initial version that has not yet been formally audited.
// A fully audited version will be published as a subsequent release.
// Use in production at your own risk.
//
// Package commondao provides governance primitives following the Common
// DAO Spec (docs/CONSTITUTION.md, Appendix): a CommonDAO is a Council (a
// set of addresses with equal voting power), a proposal lifecycle decided
// by the constitution's default voting rules, and an optional sub-DAO
// tree.
//
// Proposal types are registered per DAO: a ProposalKind couples a
// registry name with a definition factory New(dao ReadonlyCommonDAO,
// args), and Propose(creator, kind, args) accepts exactly the kinds
// registered on the DAO. New receives only a readonly view, so it cannot
// mutate the DAO before the vote; a kind that must mutate state on
// execution captures its target *CommonDAO from args (populated only by
// trusted callers) and mutates in its executor. RegisterKind /
// DeregisterKind / HasKind / KindNames and the WithProposalKind option
// are the registry primitives; the package ships one concrete kind,
// ExecutionKind (arbitrary execution), and no governance meta-kinds —
// managing a DAO's kind set through governance is the consuming realm's
// job.
//
// Proposals snapshot the council as their electorate at creation and
// are decided the moment the outcome is mathematically settled: with
// integer math over D = |electorate| - abstains, a supermajority
// (3*yes >= 2*D) passes, a NO majority (2*no > D) dismisses, and
// proposals still undecided at their voting deadline are dismissed.
//
// A DAO may carry a treasury address and frozen flag; the package only
// stores them - hosting realms derive the address and move the funds.
//
// A *CommonDAO is a mutable handle for the realm that owns it: never
// accept one from, or return one to, an untrusted realm - readonly views
// (CommonDAO.Readonly) are the only safe handles to cross a realm
// boundary. See the package README for details.
package commondao
