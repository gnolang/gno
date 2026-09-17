# gnopreview

Builds a **static gnoweb snapshot of what a pull request changed**, so a reviewer can
look at a realm instead of checking the branch out and running `gnodev`.

It is what `.github/workflows/pr-preview.yml` runs. Nothing here talks to a chain:
`gnodev` boots a fresh local node with only the selected packages loaded, the realms
render their post-`init` state, and the resulting HTML is crawled into a self-contained
directory tree.

## What gets previewed

A pull request is previewed when it touches any of:

| Change | What is rendered |
|---|---|
| a realm under `examples/gno.land/r/**` | that realm |
| a package under `examples/gno.land/p/**` | every realm that **transitively imports** it |
| `gno.land/pkg/gnoweb/**`, `gno.land/cmd/gnoweb/**`, `contribs/gnodev/**` | a fixed sample of realms, plus screenshots |

Anything else produces no preview, no artifact and **no pull request comment**.

Deliberately excluded, because they cannot change what a realm renders:
`*_test.gno`, `*_filetest.gno`, `filetests/`, `examples/quarantined/**`, and packages
marked `ignore = true`. Realms under `gno.land/r/tests/` are previewed when changed
directly but never pulled in as dependents — they are VM fixtures, and a change to a
widely imported package would otherwise crowd out the realms the PR is about.

At most 25 realms are rendered (`-max-realms`). Directly changed realms are always
kept; the comment says how many were dropped.

## Usage

```sh
# a matching gnodev — the released binary may be older than your tree
( cd contribs/gnodev && go build -o /tmp/gnodev . )

cd misc/gnopreview && go build -o /tmp/gnopreview .

# what would this branch preview?
git diff --name-only origin/master... | /tmp/gnopreview plan -changed -

# render it
git diff --name-only origin/master... > /tmp/changed.txt
GNODEV=/tmp/gnodev GNOROOT=$(git rev-parse --show-toplevel) \
  /tmp/gnopreview render -changed /tmp/changed.txt -out _preview

( cd _preview && python3 -m http.server 8777 )   # http://localhost:8777/
```

`render` writes, next to the site:

- `preview.json` — the plan that was executed, for the CI to read
- `comment.md` — the sticky pull request comment body (absent when there is nothing
  to preview, which is what makes the CI skip silently)
- `_shots/*.png` — screenshots, only for gnoweb changes

## Output layout

```
_preview/
  index.html                                  # list of rendered realms
  public/                                     # gnoweb css/js/fonts, copied from the repo
  r/gnoland/home/index.html                   # the render view
  r/gnoland/home/_t/source/index.html         # $source
  r/gnoland/home/_t/file-home.gno-source/…    # $source&file=home.gno
  r/gnoland/blog/_a/p-hello/index.html        # :p/hello render arguments
  _shots/home.png
```

gnoweb puts the tab and the render arguments in the URL path (`$source&file=x`,
`:p/about`). Those spellings are slugged into `_t/` and `_a/` so **no output path ever
contains `$`, `:` or `&`** — characters a URL tolerates but a static host may not.
Because the site lives under `_t`/`_a`/`_shots`, a `.nojekyll` marker is mandatory
wherever it is published.

Every absolute URL is rewritten: to a relative path when the target was captured, to
`https://gno.land/…` when it was not, so a link that leaves the preview lands on the
live site rather than a 404. The same relativizing is applied *inside* the copied
assets — `public/js/index.js` builds its dynamic `import()` specifiers from a
hardcoded `/public/js/controller-` prefix, which would 404 from any mount point other
than the site root, taking every interactive control with it.

## Bounds

The crawl follows links only within the selected realms, and skips URL shapes that
enumerate rather than describe:

- `$state&oid=…` / `&tid=…` — the state explorer walks the whole object graph, and
  gnoweb rate-limits it (HTTP 429) exactly as a crawler would trip it
- `$state` — ~1.3 MB of HTML per realm, describing a chain that only ever ran `init()`
- `$help&func=X` — `$help` already lists every exported function with its form
- `$download&file=X` — raw file bytes, not a page
- `:args$source` / `:args$help` — byte-identical to the argument-free tab

A `-max-pages` cap (400) backstops the rest.

## Known limits

- Transactions, the faucet, and search do not work: there is no signer and no chain
  behind the snapshot.
- Realms render their state right after `init()`. A realm that is only interesting
  once it holds data will look empty.
- Cross-realm navigation out of the previewed set goes to the live site, which may
  show a *different* version of the realm than the branch.
