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
  `Group`, which `register` requires. The qualifiers become one Go table,
  `filterKeywords`, with what a selector has (hint, label, help, scope, group,
  source) and where their values come from: a fixed set (`is:`) or a list the
  bar already holds (namespaces for `author:`, paths for `in:`). `isNarrowing`
  and `filterTakes` read it. The group headings are one Go list,
  `keywordGroups` ("Find text", "Look up", "Narrow results"). The JSON
  answer carries all three, and the popover is built from the same ones, so
  the omnibar holds no keyword or value source of its own, and of the headings
  only "On this package" and "Keywords".
- Focusing an untouched bar (empty, the prefilled path, or a results page's
  query) opens the keyword help: on a package page its keywords first, those
  that run alone on top, then one row of each group that works anywhere, so
  no group drops out; elsewhere every keyword that works anywhere, under its
  heading. The results page's popover has the same order and label on a
  package page. A header gives the keys, a last row links `/$search`, which
  on an empty query lists the keywords in the page instead of behind its
  button. `/` opens the same help, or a search of what the bar holds.
- Which page shows a package is the server's to say: `HeaderData.PackagePath`,
  set once the status and mode are known, so empty on a listing or an error
  page, whose URL has the same shape, and on an alias, whose search stays
  chain-wide. A results page goes by its URL, as it knows no more without a
  query.
- Typing the start of a keyword offers up to four keywords, from the first
  character; after `key:`, what to type and the values known locally. Choosing
  a keyword that takes a value inserts `name:`; a value completes and
  searches. A keyword that takes none runs at once only with no other word,
  as the server runs it; beside other words it is inserted. Tab
  inserts the highlighted row, and with none leaves the bar. A choice replaces
  the token its list was drawn for; in the help it adds to a query the bar
  holds, at the caret or after a query focus selected, and replaces only a
  prefilled path.
- The bar refuses nothing itself. A keyword with no value yet sends no query,
  as what it takes is the answer to give; anything else goes to the server,
  which answers or says why not. An earlier version mirrored
  discover's floor in the controller, and refused what the server answers:
  `func:` before the selector probe, `imports:` with no value, a mistyped key,
  whose name the server reports. Discover refuses before it lists anything,
  so the request costs no node query.
- Rows are `<button role="option">`s on a shared grid (keyword, description,
  marks), so descriptions align. A keyword's indexer provenance is its row's,
  as its group mixes sources: the pill in the marks column, spoken before its
  name, in the omnibar and the results page's popover alike. An answer's
  provenance is its group's, said once in the label and spoken with its name,
  so a group with no rows still says it, and 20 rows do not repeat it. The
  pointer moves the same highlight as the arrows, and leaving the list drops
  it, so Enter and Tab never act on the last row crossed.

## Alternatives considered

- Add the qualifiers to `selectors`: a selector chooses what is searched and
  one runs; a filter narrows any of them.
- Hardcode the keywords in the controller: two lists that drift.
- A value-suggestion endpoint: a request per keystroke for what results show.
- Show the help on page load: it would cover the page.

## Consequences

- The JSON answer gains `filters`, `keyword_groups`, and each selector
  `help` and `group`. A selector without `help`, from an older server during
  a rolling deploy, is still accepted.
- No JS test harness exists for the omnibar, so the controller is not
  covered in CI; it was checked in a headless browser against onyx, with and
  without an indexer.
