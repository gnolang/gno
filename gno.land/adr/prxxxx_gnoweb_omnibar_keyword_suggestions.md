# ADR: gnoweb omnibar suggests its search keywords

## Context

The omnibar understands search keywords: selectors such as `account:`,
`func:` or `deploys`, and the narrowing qualifiers `author:`, `in:` and
`is:`. The server owns the selector list and sends it with every `$search`
JSON answer, and the omnibar loaded it on first focus, but only used it to
decide what to send to the server. Nothing in the bar said keywords exist.
The only list was the "Keywords" popover on the results page, which a reader
reaches after a search, not before.

The narrowing qualifiers were not in the JSON answer at all: the popover
printed them from three hardcoded lines of its template.

## Decision

- The narrowing qualifiers become one Go table, `filterKeywords`, with a
  hint, a label, and where their values come from: a fixed set (`is:` takes
  `realm` or `package`) or a list the omnibar already holds (`namespace` for
  `author:`, `path` for `in:`). `isNarrowing` reads that table, the popover
  ranges over it, and the JSON answer carries it as `filters`. The omnibar
  names no keyword and no value source in its own code; a deployment without
  an indexer advertises fewer keywords, and the bar offers fewer.
- Focusing an untouched bar (empty, or still holding the prefilled path)
  opens the keyword help: on a realm or package page, the package's keywords,
  those that run alone first, under "On this package"; then up to four that
  work anywhere; elsewhere, every keyword that works anywhere. A last row
  links `/$search`, whose popover lists them all and whose rail has examples.
  The `/` shortcut now opens the same help instead of a path search for the
  prefilled path.
- Typing: when the word under the caret starts a keyword name, up to four
  keywords lead the dropdown under "Keywords", the typed part marked, indexer
  ones tagged and spoken as the indexer groups' rows are. A path-like word
  (`isPathLike`) offers none. Package keywords appear only on a package page
  or once `in:` is typed, where they can answer.
- After `key:`: the keyword's label heads a section that says what to type,
  read from the hint (`account:<address>` gives "Type the address"), and the
  values the bar knows locally, by prefix: `is:` from the fixed set,
  `author:` from the namespaces of the path list it already fetches, `in:`
  from that list's paths. A keyword with an empty value is not sent to the
  server: `is:` alone is refused there, `author:` alone matches nothing.
- Choosing (click, Enter on the highlighted row): a keyword that takes a
  value replaces the word under the caret with `name:` and keeps the caret
  there; a keyword that runs alone goes to the results page, as Enter on a
  qualified query does; a value completes `key:value ` and searches. Tab
  inserts the highlighted choice, or the first one typing produced, and for
  a keyword that runs alone it inserts the name rather than running it. Tab
  leaves the bar as before when the help is open and nothing is highlighted.
- Suggestions are `<button role="option">` rows built by the existing
  `buildItem`, so they join the arrow-key list, `aria-activedescendant` and
  the provenance tags unchanged. A `mousedown` on the listbox no longer
  takes focus from the input, which keeps the caret where the next keystroke
  goes and is what the combobox pattern asks for.

## Alternatives considered

- Add the narrowing qualifiers to `selectors`. A selector chooses what is
  searched and exactly one runs; a filter narrows any of them. Mixing them
  would make the omnibar's `isQualified` and `highlightTerm` reason about
  rows that never run.
- Hardcode `author`, `in` and `is` in the controller, which is shorter. The
  popover and the omnibar would then be two lists that can drift, which is
  what the selector list was sent to the browser to prevent.
- A server endpoint for value suggestions (functions of a package for
  `func:`, recent accounts for `account:`). It costs a request per keystroke
  for what the results already show; the local lists cover the qualifiers
  whose values are names a reader has to guess.
- Show the help on every empty input rather than on focus only. Clearing
  the input does reopen it; opening it on page load would cover the page.

## Consequences

- The JSON answer gains `filters`; the `is:` label changes from "only
  realms, or is:package" to "Only realms, or only packages", and the filter
  labels are capitalised like the selectors'.
- `in:` value suggestions repeat the scoped discovery rows below them: the
  first completes the query, the second navigates.
- The `/` shortcut no longer runs a path search for the prefilled path.
- No JS test harness exists for the omnibar; the behaviour is covered by a
  Playwright script run against onyx, with and without an indexer.
