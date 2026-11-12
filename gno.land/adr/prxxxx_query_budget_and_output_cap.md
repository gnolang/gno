# ADR: Node-Local Query Budgets, One Response Cap, And A Query-Only Deadline

## Context

`vm/qeval`, `vm/qrender` and the other VM queries are free, so the node pays
for whatever they do. Three costs are not covered by the gas and allocation
meters that bound VM execution. Rendering the result (`TypedValue.String`,
type text, amino JSON) runs after the machine is done and charges nothing, and
a shared value or type renders once per reference, so a few KB of input can
expand to gigabytes. The export copy that the JSON endpoints marshal is bounded
only by an estimate. And nothing bounds wall-clock time: gas counts work that
is priced, not work that is mispriced or unpriced.

## Decision

Query gas and allocation budgets become node configuration
(`application.query_max_gas`, `application.query_max_alloc`). The defaults
stay at the former 3e9 gas and 1.5 GB until production query workloads are
measured; an operator can lower them. They are not consensus state.

The export estimate and the response cap are separate limits. The 10 MB
estimate only guards the export copy; it over-counts small elements, so using
it as the response cap would refuse small answers. The 1 MiB cap is applied
once, in `vmHandler.Query` after dispatch, to every endpoint's final bytes,
envelopes included. Renderers that amplify (qeval's `WriteStringTo`, qfuncs
type text, qtype_json, amino JSON) write into a capped writer so they stop as
soon as the cap is reached instead of building the whole answer first. The
streamed text is byte-identical to `String()`, which matters because `MsgCall`
result data is rendered the same way on `DeliverTx`. The composite
`Type.String()` methods render through the same `writeType`, so type text has
one printer rather than two that must be kept equal.

A cooperative deadline stops custom-route queries at the next gas charge after
`2 × query_max_gas` nanoseconds (gas is calibrated at about 1 ns per unit), or
after `application.query_timeout` when set. It panics with its own
`DeadlineExceededError`, which realm code cannot recover and callers map to
`QueryTimeoutError` by type rather than by message. It is armed only in
`handleQueryCustom`, so `.app/simulate`, `CheckTx` and `DeliverTx` never see
it. Simulate is already bounded by the transaction's gas wanted, and stopping
it on the clock would make clients refuse transactions that fit in a block; an
operator who does not want to serve simulations can turn them off with
`rpc.allow_code_execution` (gnolang/gno#6413).

## Alternatives

A concurrency limit on heavy queries would bound their aggregate cost, but to
help it has to sit inside the proxy's existing query limiter rather than after
it, so it is left to a separate change. Streaming the export walk straight into
the JSON encoder would remove the estimate, but amino has no hook for it and a
byte mismatch would change the wire format. Reporting the deadline as an
out-of-gas error would reuse existing recover paths, but callers would then
have to tell the two apart from the message, and a gas figure cut short by the
clock reads as a real gas requirement.

## Consequences

The deadline only fires at charge points. Work that charges nothing, mainly
`TypeID` and type-text construction that `DeliverTx` shares, is not bounded
here and needs a consensus change of its own. Answers over 1 MiB on endpoints
that had no cap before (qpkg_json, qobject_*) are now refused. qpaths and
qinertpaths instead answer the leading whole paths that fit, since anyone can
deploy enough long paths to push a listing past the cap; gnoweb reports a
listing within one maximum-length path of the cap as truncated.
