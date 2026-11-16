# Harden gnoweb URL scheme checks against entity references

## Context

Gnoweb checked markdown link destinations with `html.IsDangerousURL` before
goldmark resolved HTML entity references. A destination such as
`&#x6a;avascript:` therefore passed the check and was later rendered as
`javascript:`. Leading entity-encoded C0 controls could also hide a dangerous
scheme from goldmark's prefix check.

The Gno markdown sanitizer percent-encoded unsafe URL bytes but preserved every
`&`, and its block sanitizer preserved inline-link destinations verbatim. Those
paths therefore did not provide an independent boundary against entity-based
scheme confusion.

The dangerous-scheme check was not the only consumer of the raw bytes. The link
classifier parsed the unresolved destination, so `&#x68;ttps://evil.example`
was typed as a relative path and rendered without `rel="noopener nofollow ugc"`
and without the external-link icon, while the `href` still resolved to a working
foreign URL — a reader lost the only signal that the link leaves gno.land. The
image validator (`allowSvgDataImage`) read the raw bytes as well, so
`&#x64;ata:image/png;…` passed a policy that exists to reject it.

## Decision

Resolve markdown escapes and entity references before checking a gnoweb link.
Strip leading C0 controls and spaces from the value used only for the scheme
check; render the original resolved destination so safe links retain their
existing escaped form.

Route every consumer of a destination through one `resolveDestination` helper —
the scheme check, the link classifier, and the image validator — so no consumer
can decide from bytes the renderer will later rewrite. A destination whose
resolved form `net/url` refuses to parse (an entity-encoded C0 control, say) is
typed invalid and loses its anchor rather than being emitted as a broken
relative `href` — but it keeps its link text. The unparsable class is wider
than the attacks that motivate it: an entity-spelled `%` in an absolute path
(`&percnt;`, `&#x25;`) resolves to a bare `%` that `net/url` rejects, and
`GnoLinkTypeInvalid` previously returned `WalkSkipChildren`, so those inputs
deleted author-visible copy from the page. Emitting no `<a>` is what makes the
destination inert; skipping the children was never part of that, so the invalid
branch now walks them and renders the text as plain inline content.

Trim leading C0 controls and spaces in the image policy too, for the same
reason the scheme check trims them: they are not part of the URL a browser
parses, but they shift a prefix, so `\tdata:image/png;…` showed
`AllowSvgDataImage` no `data:` at all and was waved through. `URLEscape`
happened to neutralize the result (the byte becomes `%09`, so the browser reads
a relative path), but a policy that only holds because of what a later stage
does is not a boundary. Both call sites now share one
`trimLeadingControlAndSpace` helper so the two cannot drift.

As defense in depth, percent-encode an ampersand only when it begins a numeric
or named entity-reference shape. Apply the same rule to inline link and image
destinations preserved by the block sanitizer. Bare query separators such as
`?x=1&y=2` remain unchanged.

Chain the three resolvers in `URLEscape`'s order, each over the previous
one's output rather than over the raw destination. They are not independent:
a numeric reference can manufacture the `&` that the entity-name pass then
consumes, so `javascript&#38;colon;` becomes `javascript&colon;` becomes
`javascript:`. Running the passes side by side over the original bytes would
miss it. On the sanitizer side the same chain is what makes encoding only the
first `&` sufficient — once `&#38;` is `%26#38;` there is no `&` left for a
later pass to find.

Match that shape through one level of backslash escaping. `UnescapePunctuations`
runs before the reference resolvers, so `&\#x6a;` and `&#x6a\;` both reach a
resolver as `&#x6a;`; matching the reference literally would let the escape hide
the reference from the sanitizer and reveal it to the renderer. Only the escaped
byte that matters is skipped — the `#` opening a numeric reference and the `;`
closing either form. `&\name;` is left bare because `\n` is not ASCII
punctuation, so it survives unescaping and no reference forms.

The backslash itself is not encoded. `PercentEncodeURL` does encode it (`%5C`),
which is why `sanitize.URL` and `sanitize.ImageURL` were never exposed to these
shapes, but the block sanitizer must preserve a destination's escapes so `\)`
still holds a balanced-paren destination together.

## Alternatives considered

- Encode every ampersand. Rejected because it breaks existing query-string and
  gnoweb web-query links.
- Exempt `&amp;` from the shape rule. It is the commonest entity in a URL and
  the one case where over-encoding yields a silently wrong link rather than a
  cosmetic diff: `?a=1&amp;b=2` becomes `?a=1%26amp;b=2`, one parameter valued
  `1&amp;b=2` instead of two. The exemption is sound against *this* chain —
  `&amp;` is resolved by the last pass, so its output `&` is never re-read, and
  `&amp;#x6a;avascript:` resolves only as far as `&#x6a;avascript:` (verified
  inert end to end). Rejected anyway, because it buys compatibility by
  depending on the number and order of a renderer's resolution passes, which is
  exactly the knowledge the shape rule exists to avoid: a renderer that
  resolves to a fixed point turns `&amp;#x6a;` back into `j`. Realms should
  write the bare `&`, which round-trips; the `.gno` doc comment now says so.
- Resolve references against the real HTML5 entity table instead of matching a
  shape. Rejected: it would pull the table into the gno stdlib for a
  defense-in-depth check, and shape-matching is the safer error direction — an
  unknown name like `&bogus;` gets encoded, which stays correct against a
  renderer whose table differs from ours.
- Rely only on the gnoweb renderer. Rejected because sanitized Gno output should
  not preserve an entity sequence that a later markdown renderer can resolve.
- Allowlist schemes in the renderer. Rejected as a broader compatibility change;
  this fix retains goldmark's existing dangerous-scheme policy.
- Resolve once in the AST transformer and store the result on the node.
  Rejected: `GnoLink.Destination` is goldmark's own field and other goldmark
  code reads it, so rewriting it in place would make the node disagree with the
  source. Resolving at each consumer keeps the AST faithful.

## Consequences

Entity-obfuscated `javascript:`, `data:`, `vbscript:`, and `file:` links render
with an empty `href`, including leading-control variants. Sanitized link
destinations expose entity references literally through `%26`, while ordinary
bare ampersands and safe URL rendering remain compatible.

An entity-encoded scheme now classifies exactly as its plain spelling does, so
foreign links carry `rel="noopener nofollow ugc"` and the external-link icon,
and blocked destinations render as an empty `href` with the same chrome a plain
`javascript:` link already had. Destinations that resolve to an unparsable URL
change from a broken relative `href` to no anchor at all, with the link text
still rendered — two golden files move to show that text reappearing
(`ext_link/invalid.md`, `sanitize/url-embedded-crlf`), which is the behavior
change to review.

Three `golden/sanitize` fixtures pin the fix end to end, through the real
`sanitize.X` gno helper into the gnoweb renderer, per `SANITIZE.md`'s rule that
every finding becomes a permanent regression test: an entity-encoded scheme in
a link destination, the same in an image destination, and — most load-bearing —
the backslash residual below, whose `href` is empty only because the renderer
resolves escapes before checking. Reverting `resolveDestination` turns that
last one red.

The image policy moved from a closure inside `NewDefaultGoldmarkOptions` to
`markdown.AllowSvgDataImage`, next to the validator it configures. A closure
cannot be reached from the validator's own tests, which left the policy
asserted only against a hand-copied duplicate — reverting the fix kept the
whole `gno.land` suite green. One exported definition makes those assertions
load-bearing.

`chain/markdown`'s `.gno` doc comment gained the new encoding rule, and
stdlib `.gno` source bytes are genesis state, so `expectedCrossrealm38Hash`
in `gno.land/pkg/sdk/vm` moves with it even though the rule itself lives in
the injected Go implementation.

The gnoweb image policy (`AllowSvgDataImage`) now matches case-insensitively.
Schemes and data-URI media types both are, and goldmark's `IsDangerousURL`
already compared that way, so the previous case-sensitive prefix let
`datA:imAge/png;…` show no `data:` prefix to gnoweb while goldmark still
permitted it as `data:image/png`. That let the raster URI the rule exists to
reject reach the browser without any entity encoding at all.

## Known residual

The block sanitizer still preserves backslash-escaped punctuation inside a
destination, so `[x](javascript\:alert(1))` survives `Block`/`BlockRich`
verbatim and a renderer's `UnescapePunctuations` turns it back into
`javascript:`. That is the same hazard class as an entity reference, and it is
left in place deliberately: encoding the backslash would break `\)`, which is
what holds a balanced-paren destination together. The boundary that stops it is
the gnoweb renderer, which resolves escapes before checking the scheme — the
gno-side encoding closes the entity half of the defense-in-depth layer, not the
backslash half. A downstream renderer that checks a destination without
resolving escapes first would still be exposed, exactly as it would be by any
CommonMark backslash escape.
