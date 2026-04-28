# Hardening procedure

This package (`tm2/pkg/amino/genproto`, "genproto1") generates `pbbindings.go`
files and is currently DEPRECATED for the gno.land project (see
`misc/genproto/genproto.go` for the deprecation banner). It remains in-tree
because future projects that need protobuf3-compatible generated code may
want to revive it.

If genproto1 is ever revived, it must be hardened against the same wire-format
divergence class that genproto2 was hardened against. This file documents the
systematic procedure that was used to harden genproto2 (commits on branch
`fix/jae/pr-5569`, ~12 commits resolving items #1–#26 plus follow-ups). Apply
the same procedure here.

## What the procedure produces

Byte-for-byte parity between this generator's output and the reflect codec
(`tm2/pkg/amino/binary_{encode,decode}.go`) for every registered type.
Equivalent: every conditional in the reflect path is mirrored at the
corresponding generator emission site.

## The procedure

### 1. Audit

Walk every conditional in `binary_encode.go`, `binary_decode.go`, and
`amino.go`'s Any helpers. For each conditional, ask: does the generator
emit code that honors this branch? Produce a checklist (the genproto2
audit was ~349 items in `BINARY_AUDIT.md` — that file is intentionally
untracked but lives in-repo as a reference).

For each item, classify:

- **CRITICAL** — wire-bytes diverge; reachable via registered types.
- **LATENT** — would diverge if reachable; not reachable today
  (e.g. unreachable code path, or filtered by `ValidateBasic`/zeroCheck).
- **BRITTLE** — wire-bytes match today but the structure depends on
  downstream invariants; a future refactor could regress.
- **DOC** — divergence is documented or intentional.

### 2. Per-item loop

For each CRITICAL / LATENT / BRITTLE item:

1. **Write a failing test.** Ideally re-use the parity fixture system in
   `tm2/pkg/amino/tests` (TestCodecStruct, TestCodecParity_*). If the
   bug is unreachable through the existing corpus, either:
   - Add a new test type to the corpus that exercises the divergent
     pattern (preferred — survives as a regression guard).
   - Write a one-off generator-source-inspection test (acceptable for
     pure structural fixes that have no behavioral signature today).

2. **Verify the test fails on pre-fix code.** Either run it before
   editing the generator, or temporarily revert the fix and re-run.
   If the test passes pre-fix, it isn't really testing the divergence.

3. **Apply the fix.** Edit the generator (`bindings.go` and friends in
   genproto1; `gen_marshal.go` / `gen_size.go` / `gen_unmarshal.go` in
   genproto2). Mirror the reflect-side semantics exactly — link to the
   line number in the comment.

4. **Regenerate** all `pbbindings.go` (or `pb3_gen.go` for genproto2).
   Confirm the build is clean.

5. **Run regression tests.** At minimum: `go test ./tm2/pkg/amino/...`.
   For gno.land integration: also `go test ./gno.land/pkg/integration/`,
   `go test ./gnovm/pkg/gnolang/`. (Per the iteration-pace preference,
   integration tests can be batched at the end of a fix run rather than
   per-fix.)

6. **Subagent review.** Dispatch a subagent (general-purpose or Explore)
   to independently verify the fix. Brief it with: the item description,
   the reflect ground truth, the applied fix, and what to verify
   (correctness of the fix, completeness of coverage, regression risk).
   For ambiguous items, dispatch TWO subagents with different angles
   (e.g. semantic correctness vs. empirical byte-parity).

7. **Commit.** One item per commit. Commit message references the item
   number, the reflect ground truth, the fix, and any test additions.

### 3. Cross-check signals

Watch for these during the loop:

- **A "LATENT" item turns out to be reachable.** Item #6 (this audit) was
  documented LATENT — adding `[]ReprElem7` to the test corpus immediately
  produced compile errors AND a reflect self-roundtrip failure. Always
  try to construct a reachability proof before accepting "unreachable".

- **The fix surfaces a deeper bug.** Item #6's generator fix exposed a
  pre-existing bug in `codec.go`'s `UnpackedList` determination, which
  required a reflect-side fix. Don't suppress symptoms in the generator
  if the root cause is upstream.

- **A fix's defensive branch becomes dead code.** Items #6, #19's fixes
  rendered prior defensive branches unreachable. Convert dead branches
  to explicit `panic("unreachable: ...")` rather than leaving them silent
  fallbacks — a future regression should fail loudly.

- **Sibling sites.** Most fixes have N≥2 sibling sites (encode, size,
  decode, or amino.go pre/post-decode). Fix all siblings together.

### 4. Documentation hygiene

Each fix's commit message should link to:
- The item number (e.g. `BINARY_FIXES #N`).
- The reflect ground-truth line (e.g. `binary_encode.go:592`).
- The sibling sites covered (or explicitly NOT covered, with reason).

Each fix's code change should include a comment at the fix site
referencing the reflect ground truth, so a future reader can audit the
mirror without re-deriving the contract.

## Items to prioritize for genproto1

If genproto1 is revived, start with the equivalents of these critical
genproto2 items (most likely to be reachable):

1. **fnum monotonic check** (`<` → `<=`) at every per-field decode loop.
2. **BinFixed64/32 dispatch** in any primitive-decode helper — must
   honor `fopts.BinFixed*` even at top-level repr handling.
3. **Trailing-bytes check** at every top-level non-struct unmarshal exit.
4. **`bz` slide** after every primitive decode (no `_, n, err`).
5. **`ertIsStruct` keyed off `einfo.Type.Kind()`** (the Go-side kind),
   not the repr kind, for `nil_elements` rule enforcement.
6. **Byte-array length check** in any `[N]byte` decode helper.
7. **Unpacked-list array short-input rejection** for `[N]T` array fields.
8. **Absent-field reset** in struct-level UnmarshalBinary (`*goo = T{}`
   at the top of the unmarshal body).
9. **`Any` envelope single-`0x00` rollback** at every Any-emission site
   and the matching `Size` arithmetic.
10. **`IsASCIIText(typeURL)` check** in every Any-decode entry.
11. **Post-decode `AssignableTo` re-check** before `rv.Set(irvSet)`.
12. **Float emission gated by `fopts.Unsafe`** (panic at generation time
    if not set).
13. **Uniform single-`0x00` rollback contract** at every length-prefixed
    emission site (struct-repr, Time, Duration, nested struct, interface,
    String, []byte, [N]byte, primitive — all must match
    `writeFieldIfNotEmpty:592` semantics).

The genproto2 commit log on branch `fix/jae/pr-5569` is the canonical
worked example for each of these.
