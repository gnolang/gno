// v0 - Unaudited: This is an initial version that has not yet been formally audited.
// A fully audited version will be published as a subsequent release.
// Use in production at your own risk.
//
// Package commondao provides governance primitives following the Common
// DAO Spec (docs/CONSTITUTION.md, Appendix): a CommonDAO is a Council (a
// set of addresses with equal voting power), a proposal lifecycle decided
// by the constitution's default voting rules (or custom talliers), and an
// optional sub-DAO tree.
//
// Proposals snapshot the council as their electorate at creation.
// Default-rule proposals are decided the moment the outcome is
// mathematically settled: with integer math over D = |electorate| -
// abstains, a supermajority (3*yes >= 2*D) passes, a NO majority
// (2*no > D) dismisses, and proposals still undecided at their voting
// deadline are dismissed. Custom talliers are evaluated once, at the
// deadline.
//
// A DAO may carry a treasury address and frozen flag; the package only
// stores them - hosting realms derive the address and move the funds.
//
// A *CommonDAO is a mutable handle for the realm that owns it: never
// accept one from, or return one to, an untrusted realm - readonly views
// (CommonDAO.Readonly) are the only safe handles to cross a realm
// boundary. See the package README for details.
package commondao
