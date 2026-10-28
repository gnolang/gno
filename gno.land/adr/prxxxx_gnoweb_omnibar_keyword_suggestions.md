# ADR: gnoweb omnibar suggests its search keywords

## Context

The omnibar understands search keywords: selectors such as `account:`,
`func:` or `deploys`, and the narrowing qualifiers `author:`, `in:` and
`is:`. The server sends the selector list with every `$search` JSON answer,
but the omnibar only used it to decide what to send. Nothing in the bar said
keywords exist; the only list was the results page's "Keywords" popover, and
it printed the qualifiers from three hardcoded template lines.

## Decision

- Each selector gains `Help` (what it does) and, when it works anywhere, a
  `Group`. The qualifiers become one Go table, `filterKeywords`, with hint,
  help, group and where their values come from: a fixed set (`is:`) or a
  list the bar already holds (namespaces for `author:`, paths for `in:`).
  `isNarrowing` reads it. The group headings are one Go list,
  `keywordGroups` ("Find text", "Look up", "Narrow results"). The JSON
  answer carries all three, and the popover is built from the same ones, so
  the omnibar holds no keyword, value source or heading of its own.
- Focusing an untouched bar (empty, or the prefilled path) opens the keyword
  help: on a package page its keywords first, those that run alone on top,
  then up to four that work anywhere; elsewhere every keyword that works
  anywhere, under its heading. A header gives the keys, a last row links
  `/$search`. `/` opens the same help, or a search of what the bar holds.
- Typing the start of a keyword offers up to four keywords; after `key:`,
  what to type and the values known locally. Choosing a keyword that takes a
  value inserts `name:`; one that runs alone runs; a value completes and
  searches. Tab inserts. A choice replaces the token its list was drawn for;
  Tab completes only from a list drawn for the token under the caret, and
  not while a keystroke waits for its search.
- A query the server must refuse is not sent: a keyword with no value, or,
  with no selector to run, neither two characters of text, nor of `author:`,
  nor an `in:` package (discover's floor, mirrored). The dropdown says what
  to add. This rule names `author:` and `in:` in the controller.
- Rows are `<button role="option">`s on a shared grid (keyword, description,
  marks), so descriptions align. Indexer provenance is said once in a
  group's label when the whole group shares it, else per row, pill and
  spoken text together. The pointer moves the same highlight as the arrows.

## Alternatives considered

- Add the qualifiers to `selectors`: a selector chooses what is searched and
  one runs; a filter narrows any of them.
- Hardcode the keywords in the controller: two lists that drift.
- A value-suggestion endpoint: a request per keystroke for what results show.
- Show the help on page load: it would cover the page.

## Consequences

- The JSON answer gains `filters`, `keyword_groups`, and each selector
  `help` and `group`.
- No JS test harness exists for the omnibar, so the controller is not
  covered in CI; it was checked in a headless browser against onyx, with and
  without an indexer.
