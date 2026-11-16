# Share the gnoweb foreign block cap across nested renders

## Status

Implemented.

## The problem

The `<gno-foreign>` extension renders untrusted Markdown returned by a realm
with a separate goldmark instance. Rendering is limited to 256 foreign blocks
per `Convert` call and a nesting depth of four.

Nested rendering creates a new goldmark instance and `parser.Context`. The
nesting depth was stored on the AST node and passed into the inner context, but
the foreign block count was not. Each nesting level therefore received a new
budget of 256 blocks. In the measured worst shape within the 1 MiB input limit,
gnoweb rendered 33,024 blocks before this change and 256 after it. The block
cap was bypassed by about 129 times, while render time fell from 378 ms to
81 ms, about 4.5 times. A benign 1 MiB page rendered in 12.5 ms. The realm
rendering route has no authentication or rate limit, so this allowed CPU
exhaustion despite the block cap.

## The decision

Store the block count in one private `*foreignBlockCounter`. Stash that pointer
on each admitted `ForeignNode`, then seed nested parser contexts with the same
pointer. The 256 block budget now covers the complete nested render for one
top level `Convert` call, as the depth budget already does.

Separate top level calls remain independent. `RenderRealm` creates a new
`parser.Context` for each call, and `Open` creates a counter when none exists.

Keep the existing check order. `Open` checks the block cap, checks the depth
cap, then increments the block count only after both checks succeed. An opener
rejected for depth therefore does not consume the block budget.

## Alternatives considered

**Keep a separate block count in every nested parser context.** This preserves
the previous design, but makes the 256 block cap ineffective against nested
input.

**Lower the input size cap.** This would reject legitimate large pages while
leaving the nesting amplification ratio unchanged.

**Use an exported counter field and several helper functions.** A private
shared pointer is sufficient for the parser and renderer in this package. It
avoids exported symbols and unused helper code.

## Consequences

- Production code changes only in `gno.land/pkg/gnoweb/markdown/ext_foreign.go`.
- Regression coverage changes two test files. It covers the shared nested cap,
  reset between top level calls, and ignored children beneath a stripped
  parent.
- Normal pages within both limits render as before. Blocks beyond the shared
  limit of 256 are rendered as stripped nodes.
- The app hash and gas costs do not change. gnoweb is Go code compiled into the
  node and is not on chain state.
