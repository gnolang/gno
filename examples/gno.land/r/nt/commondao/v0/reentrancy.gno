package commondao

// executing is this realm's GLOBAL re-entrancy latch, the outer of two
// complementary layers. It is set while a proposal executor is running
// (inside the public Execute) and rejects any nested Execute — an executor
// must never trigger another execution, whether of the same DAO or a
// different one, directly or by crossing back in as its DAO.
//
// The two layers guard different invariants and neither subsumes the other:
//
//   - The /p/ CommonDAO carries per-DAO `executing` / `proposing` FIELDS
//     (see the package's commondao.gno). Being fields on the owning DAO
//     object, they are writable where a mutable /p/ global is not, so they
//     travel with a standalone /p/ consumer that never wrote a realm-global
//     latch: they give every consumer same-DAO re-entrancy protection (a
//     nested Execute or Propose on the SAME DAO).
//   - This realm-global bool adds what a per-DAO field structurally cannot:
//     it enforces one-executor-per-transaction across ALL DAOs, closing the
//     cross-DAO straddle a per-DAO latch misses — the ancestor-hosted
//     dissolution/clawback case, where an executor running under DAO A's
//     Execute nests an Execute that deletes a DIFFERENT DAO out from under
//     A. A latch keyed on the executing DAO never fires there; a target
//     latch would demand a scattered check at every DAO-mutating executor.
//     The single global bool makes "one executor per tx" a local,
//     unforgettable invariant instead, and also shuts the conditional-
//     execution "free option" (nest a different DAO's Execute, observe
//     in-tx, panic to revert).
//
// It is realm module state: a mutable /p/ package global is forbidden
// (writing one panics the borrow/stamping rule), and the latch must be
// writable. This global is the realm's, complementing — not replacing — the
// /p/ per-DAO fields.
//
// Vote / Create* / Withdraw / Resign are deliberately NOT latched: no executor
// re-enters them, legitimate executors need them (sub-DAO creation, the
// dissolution sweep), and a proposal executor acting as its DAO in ANOTHER DAO
// (e.g. casting a council vote) rides those paths.
//
// A rejected nested Execute panics. Thrown across the cross() boundary the
// re-entrant call arrived through, that aborts the whole transaction.
//
// The deferred leaveExecute lowers the flag on normal returns (and on any
// same-transaction recovered panic). Cross-transaction safety does NOT rely on
// that defer — a rejected re-entry aborts the transaction, and the aborted
// transaction's realm writes (including this flag) are never committed, so a
// later transaction can never observe it stuck true.
var executing bool

// enterExecute raises the re-entrancy latch, panicking if it is already raised
// (i.e. this Execute is nested inside another). Pair with a deferred
// leaveExecute.
func enterExecute() {
	if executing {
		panic("commondao: re-entrant Execute is not allowed")
	}
	executing = true
}

// leaveExecute lowers the re-entrancy latch.
func leaveExecute() {
	executing = false
}
