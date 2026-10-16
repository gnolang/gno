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
   - Top-level fences (indented 0–3 spaces) remain open until a valid closing fence (`isCloseFence`) or EOF, auto-closing with fence characters at EOF.
2. **List-Item Fence Termination:**
   - In Rich mode (`EscapeBlockHazardsRich`), track container list indentation (`curListIndent`). When a fence opens at or beyond the list indentation, mark it as `fenceInList`.
   - If an unindented non-blank line (`isNonBlankIndentedLess`) occurs while `fenceInList` is active, terminate fence mode (`inFence = false`) without emitting a closing fence (Goldmark already closed the container). The unindented line is processed at document level, escaping any dangerous `<gno-…>` tags.
   - In strict mode (`EscapeBlockHazards`), list markers are escaped into text paragraphs, so all fences operate at document level and close only upon explicit fence closers or EOF.
3. **Shared Container Prefix Handling & `<gno-icon />` Policy:**
   - Implement `parseContainerPrefix` to scan leading indentation and CommonMark container markers (`>`, `- `, `* `, `+ `, ordered digits).
   - In `escapeGnoExtDelimiters`, escape any unescaped `<gno-…>` or `</gno-…>` at block start.
   - Allow `<gno-icon ...>` openers to remain unescaped when preceded by container markers (`hasContainer == true`) or mid-line. Keep `<gno-icon ...>` at line start (no container) escaped to maintain parity with existing golden tests.
4. **Consistent Backslash Placement:**
   - Always insert the escaping `\` after leading indentation and container markers directly before `<`, e.g. `    \<gno-card>`, rendering safely as an indented code block containing literal text.

## Alternatives Considered

- **Tracking only column-0 fences (`d9a07e79a`):** Rejected because CommonMark permits top-level code fences indented 1–3 spaces. Ignoring them opened critical bypasses for `<gno-columns>` and caused page chrome swallowing.
- **Emitting synthetic closing fences on container termination:** Rejected because CommonMark/Goldmark automatically terminates nested containers when unindented lines appear; emitting extra closing fences introduced duplicate fence lines and empty code blocks into output.
- **Exempting `<gno-icon>` at line start:** Rejected because line-start `<gno-icon>` was established and pinned as escaped in existing tests (`block-gno-icon-line-start.txtar`), and allowing it at line-start would diverge from document-level `<gno-…>` line rules.

## Consequences & Upgrade Coordination

- **Security:** Completely neutralizes `<gno-form>`, `<gno-columns>`, `<gno-frame>`, and future structural extension tags inside nested lists and containers.
- **Integrity:** Untrusted user input cannot swallow page chrome, footer content, or realm buttons via unclosed or indented fences.
- **Chain Upgrade Requirement:** `gnovm/stdlibs/chain/markdown` is an on-chain native stdlib executed deterministically by GnoVM. Changes to `EscapeBlockHazards` and `EscapeBlockHazardsRich` modify contract execution output. This change requires a coordinated protocol upgrade on `gno.land` (and Onyx/testnet) and cannot be deployed unilaterally to production nodes without a coordinated hard fork / scheduled upgrade height.
