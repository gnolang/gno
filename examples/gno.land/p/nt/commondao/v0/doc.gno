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
//
// # Extending commondao in your own realm
//
// This package is mostly mechanism: it ships the ExecutionKind concrete kind
// (with a default voting policy) and the registry primitives, and leaves the
// rest of governance policy — which kinds a DAO accepts, how it manages them,
// and any per-kind constraints such as a treasury freeze — to the consuming
// realm. To add a proposal type of your own:
//
//   - Author a ProposalKind: a type with Name() string and
//     New(dao ReadonlyCommonDAO, args any) (ProposalDefinition, error). Make
//     the definition Executable (Executor() returns an ExecFunc) if it
//     mutates state on approval. If its executor moves funds from a DAO other
//     than the proposal's host, have the HOST realm consume a Funded-style
//     contract (FundingDAOID() uint64) — minting a DAO sub needs the host's
//     cur, so this contract is host-consumed, not package-dispatched; define
//     it in your realm, as the reference realm does.
//   - Apply your own policy. ExecutionKind runs the closure as-is (no check
//     beyond a non-nil Fn), so if your realm has treasury constraints (e.g. a
//     freeze flag) do NOT catalog ExecutionKind directly: author your own
//     execution kind whose definition wraps the closure with a Validable
//     check (Validate() error) enforcing those constraints, so arbitrary
//     execution cannot bypass them. The reference realm does this so a frozen
//     DAO cannot drain its own treasury through an execution proposal.
//   - Seed it. The owning realm holds the DAO handle, so no proposal is
//     needed: pass commondao.New(WithProposalKind(YourKind{}), …) at
//     construction, or call dao.RegisterKind(YourKind{}) directly.
//   - Author a typed, CLI-friendly wrapper
//     CreateYourProposal(cur realm, daoID uint64, …params…): council-gate the
//     caller, build the args struct, and call Propose. This is the only
//     public entry, so the args-capture trust boundary holds.
//   - Optionally add a runtime governance toggle. If the council should
//     register/deregister kinds by vote (rather than only at construction),
//     author a manage-kinds-style ProposalKind whose executor calls
//     RegisterKind/DeregisterKind, and keep that managing kind itself
//     un-deregisterable so the DAO can always recover.
//
// Trust boundary: New receives only a ReadonlyCommonDAO, so a kind — even an
// externally authored one — cannot mutate the host at Propose time. The
// mutable *CommonDAO reaches a definition only through args, which your
// trusted wrapper populates (an external proposer cannot obtain one). On
// execution the host passes the DAO's terminal, RealmSend-only sub, so a
// fund-moving executor is bounded to that one DAO address. See the reference
// realm gno.land/r/nt/commondao/v0 for a full worked example.
package commondao
