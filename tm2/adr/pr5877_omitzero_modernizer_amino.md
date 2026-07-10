# pr5877: Disable the `omitzero` `go fix` modernizer (Amino honors `omitempty` on structs)

## Status

Accepted

## Context

PR #5877 applies the Go 1.26 `go fix` modernizers across every module CI lints
and enforces them with a `go fix -diff ./...` check in
`.github/workflows/_ci-go.yml` (and a matching `make fix`).

One of the registered modernizers, `omitzero`, "identifies uses of the
`omitempty` JSON struct tag on fields that are themselves structs" and either
removes the tag or replaces it with `omitzero`. Its premise is that for
`encoding/json`, `omitempty` has no effect on struct-typed fields, so the tag is
dead.

That premise is false for **Amino**, which is used pervasively across this
repository. Amino parses `,omitempty` from the `json` tag
(`tm2/pkg/amino/codec.go`) and, when marshaling to JSON, skips a struct-typed
field whose value equals its zero value
(`tm2/pkg/amino/json_encode.go`: `if field.JSONOmitEmpty && isJSONEmpty(...)`).
Amino does **not** recognize `omitzero`. So on an Amino-serialized type:

- removing `,omitempty` makes a previously-omitted zero struct always serialize;
- replacing `,omitempty` with `,omitzero` does the same (Amino ignores it).

Applying `omitzero` blanket stripped the tag from Amino types and changed their
JSON output. It was caught by `gno.land/pkg/integration/testdata/qeval_json.txtar`,
whose golden output for a struct query gained the zero-valued `Hash`/`OwnerID`
fields of `ObjectInfo`. The same removal also affected `RefValue` (realm state
JSON), `GenesisDoc.ConsensusParams` (genesis JSON) and `ResultTx.Proof` (RPC
JSON). Amino **binary** encoding is unaffected — it has its own default-skipping
independent of the tag — so this is a JSON-format regression, not a state-hash or
consensus break.

The modernizer already carves out this exact class of problem: it skips packages
with `+kubebuilder` annotations because kubebuilder assigns its own meaning to
the tag. Amino is in the same position, but the modernizer has no way to know it.

## Decision

Disable the `omitzero` modernizer repo-wide by passing `-omitzero=false` to
`go fix` in both the CI check (`.github/workflows/_ci-go.yml`) and `make fix`
(via `GO_FIX_FLAGS`), and restore the `,omitempty` tags it removed from
Amino-serialized types:

- `gnovm/pkg/gnolang/ownership.go` — `ObjectInfo.Hash`, `ObjectInfo.OwnerID`
- `gnovm/pkg/gnolang/values.go` — `RefValue.ObjectID`, `RefValue.Hash`
- `gnovm/pkg/gnomod/file.go` — `File.AddPkg`
- `tm2/pkg/bft/rpc/core/types/responses.go` — `ResultTx.Proof`
- `tm2/pkg/bft/types/genesis.go` — `GenesisDoc.ConsensusParams`

## Alternatives considered

- **Accept the change and update the goldens.** Rejected: it silently alters the
  JSON wire format of genesis, RPC responses and realm queries.
- **Replace `omitempty` with `omitzero`.** Rejected: Amino does not honor
  `omitzero`, so the field would still always serialize.
- **Per-field suppression.** `go fix` modernizers have no inline-ignore
  directive for `omitzero`, and adding fake `+kubebuilder` markers to trip its
  carve-out would be misleading. Disabling the one modernizer is cleaner.

## Consequences

- All other modernizers stay enforced; only `omitzero` is off.
- Genuinely dead `omitempty` tags on non-Amino, `encoding/json`-only struct
  fields will no longer be auto-flagged. This is a cosmetic loss and the safe
  trade for not corrupting Amino output.
- Future contributors adding `,omitempty` to an Amino struct field keep the
  behavior they expect; the CI check will no longer fight them.
