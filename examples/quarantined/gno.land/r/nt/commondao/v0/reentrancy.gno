package commondao

// executing is the re-entrancy latch. It is set while a proposal executor is
// running (inside Execute) and rejects any nested Execute — an executor must
// never trigger another execution, whether of the same DAO or a different one,
// directly or by crossing back in as its DAO.
//
// It is realm module state: a mutable /p/ package global is forbidden (writing
// one panics the borrow/stamping rule), and the latch must be writable. A
// single bool suffices because the invariant is global — at most one executor
// runs per transaction. Keying it per-DAO would let an executor nest an
// Execute of a different DAO (re-opening a conditional-execution "free option")
// and would demand a scattered target-latch check at every DAO-mutating
// executor; the global bool makes "one executor per tx" a structural, local,
// unforgettable invariant instead.
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
