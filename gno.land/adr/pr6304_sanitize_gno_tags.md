# ADR: Sanitize `<gno-…>` tags at block start, isolate nested fences, and allow container `<gno-icon />`

## Status

Proposed (coordinated chain upgrade required).

## Context

`p/nt/markdown/sanitize` (implemented on top of `chain/markdown` natives) neutralizes dangerous constructs in user-supplied markdown before rendering in `gnoweb`. Several vulnerabilities and inconsistencies emerged around block tags, fenced code blocks, and extension tags:

1. **Issue #6300 (Fence inside list items):** A code fence indented under a list item (e.g. `- a\n  ```\n<gno-form path="x">\n```\n`) caused the sanitizer's line loop to enter fence mode (`inFence = true`). However, CommonMark §5.2 dictates that an unindented non-blank line (`<gno-form ...>`) cannot be part of the preceding list item, terminating both the list item and its nested code block. Goldmark parsed `<gno-form>` as a document-level element, rendering a live interactive form, while the sanitizer skipped escaping. Furthermore, the trailing fence line was treated by Goldmark as opening an unclosed code block at document level, swallowing subsequent page chrome and realm footers.
2. **Indented Fences (0–3 spaces):** An initial fix attempt (commit `d9a07e79a`) restricted fence tracking to column 0 (`if i != 0`), ignoring fences indented 1–3 spaces. Under CommonMark §4.5, top-level fences may be indented 0–3 spaces. Ignoring them caused bypasses where indented fences followed by fences/tags (such as ` ~~~\n~~~\n<gno-columns>` and ` ```\nx\n```\n<gno-columns>`) caused subsequent closing fences to be misinterpreted as openers, leaving `<gno-columns>` live. Unclosed indented fences also failed to auto-close at EOF, swallowing realm footers.
3. **Container `<gno-icon />` Rendering:** `<gno-icon />` is an inline extension meant to render safely in user content. While top-level line-start tags are escaped to avoid block-level ambiguity, icons at the beginning of list items (`- <gno-icon ...>`) and blockquotes (`> <gno-icon ...>`) were inadvertently neutralized when container prefix stripping was introduced.
4. **Indentation and PR #6298 Harmony:** PR #6298 introduced backslash placement before 4+ spaces of indent for `<gno-button>`. For general extension delimiters, placing the backslash after indentation (`    \<gno-card>`) preserves indented code blocks per CommonMark while leaving `indentColumns` unused.

## Decision

1. **Unified Fence Detection across Line Loop and Bracket Walker:**
   - Detect fence openers uniformly in `escapeLineLeader` and `findBracketSpans` using `parseFenceOpener`, supporting 0–3 leading spaces and enforcing CommonMark §4.5 info-string rules (no backticks in backtick fence info strings).
   - The bracket walker runs the fence check once per line, at the line's first byte, and lets `parseFenceOpener` take the indent. Re-running it at every leading space made the walker quadratic in the indent (1 MiB of spaces took 12 s, against 5 ms before) and accepted a fence at any indent, so `    ```\n[home]: https://evil.example` kept a link reference definition that goldmark reads at document level.
   - Fences remain open until a valid closing fence (`isCloseFence`) or EOF. The EOF auto-close is written at the opener's column: at column 0 it would end a list item holding the fence and open a new fence over the realm's chrome.
2. **Fence Termination by Indent (Rich mode):**
   - A fence inside a list item ends, in goldmark, at the first non-blank line indented less than the item's content, and that line then parses at document level. Rich mode keeps list markers, so it must not leave such a line inside its fence state.
   - Rather than track list items, a pre-pass (`closeFencesAtDedent`) writes a closer at the opener's column before the first non-blank line indented less than the opener, and moves a closer that sits left of its opener to the opener's column. A closer at the opener's column ends the fence whether or not the fence sits in a list item, so goldmark and the sanitizer agree on every line after it.
   - The pre-pass runs before the bracket walker and the line loop, so neither applies an indent rule of its own. Applying it in each pass separately let them disagree: the walker deletes link reference definitions, so a definition that ended a fence by its indent was gone by the time the line loop ran, and the loop took the user's next fence line as a closer where the walker had taken it as an opener. `- a\n  ```\n[x]: u\n```\n[home]: https://evil.example` then kept the second definition live.
   - In strict mode (`EscapeBlockHazards`), list markers are escaped into text paragraphs, so all fences operate at document level and close only upon explicit fence closers or EOF.
3. **Shared Container Prefix Handling & `<gno-icon />` Policy:**
   - Implement `parseContainerPrefix` to scan leading indentation and CommonMark container markers (`>`, `- `, `* `, `+ `, ordered digits). It and `escapeLineLeader` parse list markers through one helper, `listMarkerEnd`.
   - In `escapeGnoExtDelimiters`, escape any unescaped `<gno-…>` or `</gno-…>` at block start.
   - Allow `<gno-icon ...>` openers to remain unescaped when preceded by container markers (`hasContainer == true`) or mid-line. Keep `<gno-icon ...>` at line start (no container) escaped to maintain parity with existing golden tests.
4. **Consistent Backslash Placement:**
   - Always insert the escaping `\` after leading indentation and container markers directly before `<`, e.g. `    \<gno-card>`, rendering safely as an indented code block containing literal text.

## Alternatives Considered

- **Tracking only column-0 fences (`d9a07e79a`):** Rejected because CommonMark permits top-level code fences indented 1–3 spaces. Ignoring them opened critical bypasses for `<gno-columns>` and caused page chrome swallowing.
- **Tracking list items (`curListIndent` / `fenceInList`):** An earlier revision of this PR ended a fence on a dedent only when it had seen the fence open inside a list item. A blank line (`- a\n\n  ```\n<gno-form path="x">`) and a lazy continuation line (`- a\nb\n  ```\n<gno-columns>`) both reset its list state while goldmark kept the item open, so the tag rendered live and the realm footer was lost. Empty list items (`-` alone on a line) would be a third gap. In the other direction, lines shaped like list markers that goldmark does not open a list for (a thematic break `* * *`, or `2. item` right after a paragraph) made it leave a document-level fence without a closer, so the fence ran over the realm's chrome. Modelling goldmark's list rules line by line is what #6300 calls fragile; the indent rule needs none of it.
- **Exiting fence state on a dedent without writing a closer:** Correct only when the fence really sits in a list item. For a document-level fence indented 1–3 spaces, goldmark keeps the dedented line as code and treats the user's next fence line as the closer, so the sanitizer's EOF auto-close would open a new fence over the realm's chrome. The written closer makes both readings agree.
- **Exempting `<gno-icon>` at line start:** Rejected because line-start `<gno-icon>` was established and pinned as escaped in existing tests (`block-gno-icon-line-start.txtar`), and allowing it at line-start would diverge from document-level `<gno-…>` line rules.

## Consequences & Upgrade Coordination

- **Security:** `<gno-form>`, `<gno-columns>`, `<gno-frame>`, and future structural extension tags are escaped at block start, including after list and quote markers and after a fence that a list item ends.
- **Integrity:** Unclosed, indented, and list-item fences no longer swallow the realm's chrome or rebind its reference links. `TestExtensionAndFooterIsolation` renders each known shape between realm chrome and checks for live `gno-*` elements, a bolded footer, and an unbound `[home]`.
- **Fidelity cost (Rich mode only):** A fence opened 1–3 spaces in whose content or closer sits further left than the opener is now closed at that line. ` ```\nfoo\n``` ` renders as an empty code block, a paragraph, and a second code block where goldmark alone would render one code block. Fences whose content is indented like the opener, and fences inside list items, render as before.
- **Chain Upgrade Requirement:** `gnovm/stdlibs/chain/markdown` is an on-chain native stdlib executed deterministically by GnoVM. Changes to `EscapeBlockHazards` and `EscapeBlockHazardsRich` modify contract execution output. This change requires a coordinated protocol upgrade on `gno.land` (and Onyx/testnet) and cannot be deployed unilaterally to production nodes without a coordinated hard fork / scheduled upgrade height.
