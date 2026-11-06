#!/usr/bin/env bash
#
# Fills the "What's new" block of a static home page, in place. See README.md.
# Usage: whatsnew.sh [--dry-run] [--force] [--builders] <home.md>
set -euo pipefail

: "${RPC_URL:=https://rpc.gno.land:443}"
# The blog only lives on mainnet: a testnet reads it there, and sets LINK_BASE
# to https://gno.land so that its links lead there.
: "${BLOG_RPC_URL:=https://rpc.gno.land:443}"
: "${LINK_BASE:=}"
# Without an indexer, the realms stay out of the page.
: "${INDEXER_URL:=}"
: "${X_ACCOUNT:=_gnoland}"
# Deployments whose path matches this regular expression stay out of --builders.
: "${DEPLOYED_SKIP=/smoke/}"
# For a gnoweb that reads its static pages at startup only.
: "${ON_CHANGE:=}"
# Refresh intervals, in minutes.
: "${FEED_TTL:=30}"
: "${USED_TTL:=1440}"
: "${IMPORTED_TTL:=10080}"
# The testnet --builders describes.
: "${NETWORK:=onyx}"
export RPC_URL BLOG_RPC_URL LINK_BASE INDEXER_URL X_ACCOUNT DEPLOYED_SKIP USED_NAMESPACES

START='<!-- whats-new:start -->'
END='<!-- whats-new:end -->'
TRENDING_START='<!-- whats-trending:start -->'
TRENDING_END='<!-- whats-trending:end -->'
CARD_ROWS=3
LIST_SIZE=5
SHIPPED_SIZE=3
MIN_CHARS=20
MAX_IMPORT_FAILURES=10
MAX_CHARS=110
SHORT_CHARS=60
ACTIVITY_WINDOW=604800
# Past a week, a ranking no longer reads as "this week".
USED_MAX_AGE=$((ACTIVITY_WINDOW / 60))
RATE_PROBE=20000
BAND_WIDTH=25000
MIN_BAND_WIDTH=250
MONTHS='Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec'
YMD='def ymd: .m as $m | "\(.y)\(($months | split(" ") | index($m)) + 101 | tostring | .[1:])\(.d | if length < 2 then "0" + . else . end)";'
# Only these namespaces can reach the page: others are permissionless, and
# anyone can inflate a ranking with throwaway addresses.
: "${USED_NAMESPACES=gnoland gnoswap akkadia}"

CURL=(curl -fsSL --max-time 20 -A whatsnew.sh)

abci() {
  local hex
  hex=$(printf %s "$3" | od -An -tx1 | tr -d ' \n')
  "${CURL[@]}" -G --data-urlencode "path=\"$2\"" --data-urlencode "data=0x$hex" "$1/abci_query" |
    jq -er '.result.response.ResponseBase | if .Error then error("query failed") else .Data // error("no data") | @base64d end'
}

graphql() {
  "${CURL[@]}" --max-time 60 -H 'content-type: application/json' \
    --data "$(jq -n --arg query "$1" '{query: $query}')" "$INDEXER_URL"
}

# head would close the pipe early, and pipefail would read that as a failure.
first_rows() {
  awk -v n="$1" 'NR <= n'
}

html_to_text() {
  jq -Rrs '
    gsub("<a [^>]*>pic\\.(twitter|x)\\.com/[^<]*</a>"; "")
    | gsub("<br ?/?>"; " ")
    | gsub("<[^>]*>"; "")
    | gsub("&#(?<n>[0-9]+);"; ([.n | tonumber] | implode))
    | gsub("&lt;"; "<") | gsub("&gt;"; ">") | gsub("&quot;"; "\"")
    | gsub("&nbsp;"; " ") | gsub("&amp;"; "&")'
}

# A t.co link to a bare domain becomes that domain as text; every other link is
# dropped, so that the page never carries a link a post chose.
expand_links() {
  local text url target
  text=$(cat)
  for url in $(grep -oE 'https://t\.co/[A-Za-z0-9]+' <<<"$text" | sort -u || true); do
    target=$(curl -s --max-time 10 -o /dev/null -w '%{redirect_url}' "$url" || true)
    if grep -qE '^https?://(www\.)?gno\.land(/|$)' <<<"$target"; then
      text=${text//$url/Gno.land}
    elif grep -qE '^https?://(www\.)?([a-z0-9-]+\.)+[a-z]{2,}/?$' <<<"$target"; then
      target=${target#*://}
      target=${target#www.}
      text=${text//$url/${target%/}}
    fi
  done
  jq -Rrs '
    gsub("https?://[^[:space:]]+"; "")
    | gsub("[[:cntrl:][:space:]]+"; " ")
    | gsub("^ +| +$"; "")' <<<"$text"
}

# Keeps text on one line and shortens it, then escapes all ASCII punctuation, and
# the g of a bare address: no text can form a link, image, tag, mention or icon,
# or shift the fields of a row.
to_md() {
  jq -Rrs --argjson max "$MAX_CHARS" '
    gsub("[[:cntrl:][:space:]]+"; " ")
    | gsub("[\u200E\u200F\u202A-\u202E\u2066-\u2069]"; "")
    | gsub("^ +| +$"; "")
    | if length > $max
      then .[:$max] | sub(" [^ ]*$"; "") | sub("[ ,;:.-]+$"; "") + "…"
      else . end
    | gsub("(?<c>[!-/:-@\\[-`{-~])"; "\\\(.c)")
    | gsub("g(?=1[0-9A-Za-z]{36})"; "&#103;")'
}

# A source prints one row per item: date (YYYYMMDD), link, escaped text.
row() {
  [[ $1 =~ ^[0-9]{8}$ ]] || return 1
  printf '%s\t%s\t%s\n' "$1" "$2" "$(to_md <<<"$3")"
}

# FxTwitter only supplies the post IDs: text and author come from X itself.
source_x() {
  local posts id kind url oembed date text n=0
  posts=$("${CURL[@]}" "https://api.fxtwitter.com/2/profile/${X_ACCOUNT}/statuses" |
    jq -r --arg account "$X_ACCOUNT" '
      .results[:20][]
      | select(.author.screen_name == $account
          and .reposted_by == null and .replying_to == null
          and (.id | test("^[0-9]+$")))
      | [.id, (if .article then "article" else "x" end)] | @tsv' | sort -rn)
  while IFS=$'\t' read -r id kind; do
    [ "$n" -lt "$CARD_ROWS" ] || break
    url="https://x.com/${X_ACCOUNT}/status/${id}"
    oembed=$("${CURL[@]}" -G --data-urlencode "url=${url}" --data-urlencode "omit_script=1" \
      https://publish.x.com/oembed) || continue
    [ "$(jq -r .author_url <<<"$oembed")" = "https://x.com/${X_ACCOUNT}" ] || continue
    date=$(jq -er --arg months "$MONTHS" "$YMD"'
      .html
      | capture("<a href=\"https://x\\.com/[^\"]*\">(?<m>[A-Z][a-z]{2})[a-z]* (?<d>[0-9]{1,2}), (?<y>[0-9]{4})</a></blockquote>")
      | ymd' <<<"$oembed") || continue
    if [ "$kind" = "article" ]; then
      # oEmbed only returns a link for an X Article; the title sits on X's CDN.
      text=$("${CURL[@]}" "https://cdn.syndication.twimg.com/tweet-result?id=${id}&lang=en&token=4" |
        jq -r '.article.title // empty') || continue
      [ -n "$text" ] || continue
    else
      text=$(jq -r '.html | gsub("\\s+"; " ") | capture("<p[^>]*>(?<p>.*?)</p>").p' <<<"$oembed" |
        html_to_text | expand_links)
      [ "${#text}" -ge "$MIN_CHARS" ] || continue
    fi
    row "$date" "$url" "$text"
    n=$((n + 1))
  done <<<"$posts"
  [ "$n" -gt 0 ]
}

source_blog() {
  local fields date slug title
  fields=$(abci "$BLOG_RPC_URL" vm/qrender 'gno.land/r/gnoland/blog:' |
    jq -Rrs --arg months "$MONTHS" --argjson size "$CARD_ROWS" "$YMD"'
      limit($size; capture("### \\[(?<t>[^\\n]*)\\]\\(/r/gnoland/blog:p/(?<s>[a-z0-9-]+)\\)\\n(?<d>[0-9]{2}) (?<m>[A-Z][a-z]{2}) (?<y>[0-9]{4})"; "g"))
      | [ymd, .s, (.t | gsub("[\\t\\n\\r]"; " "))] | join("\t")')
  [ -n "$fields" ]
  while IFS=$'\t' read -r date slug title; do
    row "$date" "${LINK_BASE}/r/gnoland/blog:p/${slug}" "$title"
  done <<<"$fields"
}

source_release() {
  local fields tag date
  fields=$("${CURL[@]}" 'https://api.github.com/repos/gnolang/gno/releases?per_page=30' |
    jq -er --argjson size "$LIST_SIZE" '
      [.[] | select(.prerelease == false and (.tag_name | test("^v[0-9.]+$")))][:$size][]
      | [.tag_name, (.published_at[0:10] | gsub("-"; ""))] | @tsv')
  [ -n "$fields" ]
  while IFS=$'\t' read -r tag date; do
    row "$date" "https://github.com/gnolang/gno/releases/tag/${tag}" "$tag"
  done <<<"$fields"
}

# scan_band <lower> <upper> <filter> <fields> <rows> prints the rows a jq filter
# draws from the heights (lower, upper], splitting a band the indexer refuses as
# too large.
scan_band() {
  local lower=$1 upper=$2 filter=$3 fields=$4 rows=$5 reply middle
  reply=$(graphql "{ getTransactions(where: {
      block_height: { gt: $lower, lt: $((upper + 1)) }
      messages: { value: { $filter } }
    }) { block_height success messages { value { __typename $fields } } } }") || return 1
  if ! jq -e .errors <<<"$reply" >/dev/null; then
    # tx-indexer answers null for a band with no match: an empty band, not a
    # failure, which the .errors check above already caught.
    jq -r ".data | if has(\"getTransactions\") then .getTransactions // [] | .[] else error(\"no data\") end
      | select(.success) | .block_height as \$height | .messages[]?.value | $rows" <<<"$reply"
    return
  fi
  [ $((upper - lower)) -gt "$MIN_BAND_WIDTH" ] || return 1
  middle=$(((lower + upper) / 2))
  scan_band "$lower" "$middle" "$filter" "$fields" "$rows" &&
    scan_band "$middle" "$upper" "$filter" "$fields" "$rows"
}

block_time() {
  graphql "{ getBlocks(where: { height: { eq: $1 } }) { time } }" |
    jq -er '.data.getBlocks[0].time | .[0:19] + "Z" | fromdate'
}

# Stops at the first failed band: a partial count must not pass for a full one.
scan_week() {
  local tip head probe lower upper
  tip=$(graphql '{ latestBlockHeight }' | jq -er '.data.latestBlockHeight | numbers') || return 1
  head=$(block_time "$tip") || return 1
  # A chain younger than the probe, or than the window, counts from its first
  # block, as gnoweb's windowStart does.
  lower=0
  if [ "$tip" -gt "$RATE_PROBE" ]; then
    probe=$(block_time $((tip - RATE_PROBE))) || return 1
    lower=$((tip - RATE_PROBE * ACTIVITY_WINDOW / (head - probe)))
  fi
  [ "$lower" -gt 0 ] || lower=0
  while [ "$lower" -lt "$tip" ]; do
    upper=$((lower + BAND_WIDTH))
    [ "$upper" -le "$tip" ] || upper=$tip
    scan_band "$lower" "$upper" "$@" || return 1
    lower=$upper
  done
}

# Shortens the address of a personal namespace: g1pq6p…/sinopia/gallery.
short_paths() {
  awk -F'\t' -v OFS='\t' '{
    path = substr($2, 9); name = substr(path, 4)
    if (name ~ /^g1[a-z0-9]+\//) name = substr(name, 1, 6) "…" substr(name, index(name, "/"))
    print $1, name, path
  }'
}

# Ranks by distinct addresses: raw call counts are dominated by bots. It writes
# the week's totals beside its ranking, since it has already read them.
source_used() {
  local realms="^gno.land/r/" calls
  [ -z "$USED_NAMESPACES" ] || realms="^gno.land/r/(${USED_NAMESPACES// /|})/"
  calls=$(scan_week "MsgCall: { pkg_path: { like: \"$realms\" } }" '... on MsgCall { caller pkg_path }' \
    'select(.__typename == "MsgCall" and .pkg_path != null) | [.pkg_path, .caller] | @tsv' |
    grep -E "$realms" | grep -E $'^gno\\.land/r/[a-z0-9_./-]+\t' | grep -vF .. | sort -u)
  printf '%s\t%s\n' "$(cut -f2 <<<"$calls" | sort -u | grep -c .)" "$(cut -f1 <<<"$calls" | sort -u | grep -c .)" \
    >"$STATE_DIR/activity"
  cut -f1 <<<"$calls" | sort | uniq -c | sort -rn | first_rows "$LIST_SIZE" |
    awk -v OFS='\t' '{ print $1, $2 }' | short_paths
}

# A chain can park code, so addpkg alone does not put a package online. One
# row per project (namespace/name), its newest live deploy: a newer one that
# is parked or refused must not hide an older live one. At most LIST_SIZE * 4
# deploys are checked, and only those of a project not listed yet: a project
# that deployed many packages must not crowd the others out.
live_rows() {
  local height pkg project seen=" " n=0 tries=0
  while IFS=$'\t' read -r height pkg; do
    [ "$n" -lt "$LIST_SIZE" ] && [ "$tries" -lt "$((LIST_SIZE * 4))" ] || break
    project=$(cut -d/ -f3,4 <<<"$pkg")
    case $seen in *" $project "*) continue ;; esac
    tries=$((tries + 1))
    abci "$RPC_URL" vm/qfile "$pkg" >/dev/null 2>&1 || continue
    seen="$seen$project "
    printf '%s\t%s\n' "$height" "$pkg"
    n=$((n + 1))
  done <<<"$1"
}

# One line per project, newest first; the week's count goes beside it.
source_deployed() {
  local rows height name path epoch
  rows=$(scan_week 'MsgAddPackage: { package: { path: { like: "." } } }' '... on MsgAddPackage { package { path } }' \
    'select(.__typename == "MsgAddPackage") | [$height, .package.path] | @tsv' |
    awk -F'\t' -v skip="$DEPLOYED_SKIP" '/^[0-9]+\tgno\.land\/[rp]\/[a-z0-9_.\/-]+$/ && !index($0, "..") && (skip == "" || $2 !~ skip)' |
    sort -rn)
  { grep -c . <<<"$rows" || true; } >"$STATE_DIR/deployed-count"
  live_rows "$rows" | short_paths | while IFS=$'\t' read -r height name path; do
    epoch=$(block_time "$height")
    printf '%s\t%s\t%s\n' "$(jq -nr --argjson t "$epoch" '$t | strftime("%Y%m%d")')" "$(cut -d/ -f1,2 <<<"$name")" "$path"
  done
}

# The binary the chain runs, a candidate on a testnet. node_info.version is the
# protocol's, not the release's. Only a release tag goes into the link.
source_version() {
  "${CURL[@]}" "$RPC_URL/status" |
    jq -er '.result.build_version | select(test("^v[0-9.]+(-rc\\.[0-9]+)?$"))'
}

source_shipped() {
  local prs number title
  prs=$("${CURL[@]}" 'https://api.github.com/search/issues?q=repo:gnolang/gno+is:pr+is:merged&sort=updated&order=desc&per_page=30' |
    jq -r --argjson size "$SHIPPED_SIZE" '[.items[] | select(.title | test("^(feat|fix)[(:!]"))] | sort_by(.closed_at) | reverse | .[:$size][]
      | (.title | sub("^[a-z]+(\\([^)]*\\))?!?: *"; "")) as $t
      | [(.number | tostring), (($t[:1] | ascii_upcase) + $t[1:] | gsub("[\\t\\n\\r]"; " "))] | join("\t")')
  [ -n "$prs" ]
  while IFS=$'\t' read -r number title; do
    printf '%s\t%s\n' "$number" "$(MAX_CHARS=$SHORT_CHARS to_md <<<"$title")"
  done <<<"$prs"
}

source_imported() {
  local imports
  imports=$({ abci "$RPC_URL" 'vm/qpaths?limit=10000' gno.land/p/ && abci "$RPC_URL" 'vm/qpaths?limit=10000' gno.land/r/; } |
    grep . | xargs -P 4 -I{} "$BASH" "$0" --imports {})
  # A package may fail to read; many failing would skew the ranking.
  [ "$(grep -c '^!failed$' <<<"$imports" || true)" -le "$MAX_IMPORT_FAILURES" ]
  grep -E '^gno\.land/p/[a-z0-9_./-]+$' <<<"$imports" | grep -vF .. | grep -vE '^gno\.land/p/g1[a-z0-9]{38}/' |
    sort | uniq -c | sort -rn | first_rows "$LIST_SIZE" |
    awk '{
      path = $2; sub(/^gno\.land/, "", path)
      name = path; sub(/\/v[0-9]+$/, "", name); sub(/.*\//, "", name)
      print $1 "\t" name "\t" path
    }'
}

older_than() {
  [ -n "$(find "$1" -mmin "+$2")" ]
}

# refresh <source> <ttl> runs a source at most every ttl minutes, even when it
# fails. A failure leaves the previous output in place.
refresh() {
  local fragment="$STATE_DIR/$1" stamp="$STATE_DIR/$1.tried" tmp
  if [ "$FORCE" = 0 ] && [ -f "$stamp" ] && ! older_than "$stamp" "$2"; then
    return 0
  fi
  touch "$stamp"
  tmp=$(mktemp "$fragment.XXXXXX")
  if "$BASH" "$0" --source "$1" >"$tmp" 2>"$tmp.err" && [ -s "$tmp" ]; then
    mv "$tmp" "$fragment"
  else
    rm -f "$tmp"
    echo "whatsnew: $1 failed, keeping its previous content: $(tail -n 1 "$tmp.err")" >&2
  fi
  rm -f "$tmp.err"
}

short_date() {
  local months=($MONTHS)
  printf '%s %d' "${months[$((10#${1:4:2} - 1))]}" "$((10#${1:6:2}))"
}

latest() {
  [ ! -f "$STATE_DIR/$1" ] || awk -F'\t' 'NF == 3' "$STATE_DIR/$1" | sort -rs -t$'\t' -k1,1 | first_rows "$CARD_ROWS"
}

is_recent() {
  [ -s "$STATE_DIR/$1" ] && ! older_than "$STATE_DIR/$1" "$USED_MAX_AGE"
}

# One row of the page's grid.
render_columns() {
  local column first=1
  printf '<gno-columns>\n'
  for column in "$@"; do
    [ -n "$column" ] || continue
    [ "$first" = 1 ] || printf '\n<gno-columns-sep>\n'
    printf '\n%s\n' "$column"
    first=0
  done
  printf '\n</gno-columns>\n\n'
}

# Laid out like posts: a short green link, then neutral text.
render_post() {
  local date url text
  [ -n "$3" ] || return 0
  IFS=$'\t' read -r date url text <<<"$3"
  printf '%s **[%s · %s](%s)**\n\n%s' "$1" "$2" "$(short_date "$date")" "$url" "$text"
}

render_button() {
  printf '<gno-button href="%s" label="%s" variant="outline" />' "$1" "$2"
}

# One source per row, on the page's three columns, with its button under it, so
# that a row stays level whatever the length of its posts, and a post on X every
# day cannot push the blog out.
render_row() {
  local line cells=()
  [ -n "$3" ] || return 0
  while IFS= read -r line; do
    cells+=("$(render_post "$1" "$2" "$line")")
  done <<<"$3"
  render_columns "${cells[@]}"
  printf '%s\n\n' "$4"
}

render_block() {
  local posts blog
  posts=$(latest x)
  blog=$(latest blog)
  [ -n "$posts$blog" ] || return 0
  printf '## <gno-icon name="newspaper" /> What'\''s new\n\n'
  render_row '𝕏' "Gno.land \\@${X_ACCOUNT/_/\\_}" "$posts" "$(render_button "https://x.com/${X_ACCOUNT}" 'Follow on X')"
  render_row '<gno-icon name="book" />' Blog "$blog" "$(render_button "${LINK_BASE}/blog" 'Read the blog')"
  # One break after the block: between its rows it would read as a new section.
  printf -- '---\n\n'
}

# A list and the button to all of it stay in one column, so that they stay
# together once the columns stack on a phone.
with_button() {
  [ -n "$1" ] || return 0
  printf '%s\n\n%s' "$1" "$(render_button "$2" "$3")"
}

# The lists of the page's Explore section, in one frame that sets them apart from
# the text above. Without an indexer, the merged pull requests keep them at three,
# two of them so that the column stays as long as the lists beside it.
render_trending() {
  local used imported releases shipped pulls='https://github.com/gnolang/gno/pulls?q=is%3Apr+is%3Amerged'
  used=$(render_used)
  imported=$(with_button "$(render_imported)" /p/ 'Explore packages')
  releases=$(with_button "$(render_releases)" https://github.com/gnolang/gno/releases 'All releases')
  shipped=$(with_button "$(render_shipped 2)" "$pulls" 'All merged PRs')
  [ -n "$imported$releases" ] || return 0
  printf '<gno-frame>\n\n'
  if [ -n "$used" ]; then
    render_columns "$(with_button "$used" /r/ 'Explore realms')" "$imported" "$releases"
  else
    render_columns "$imported" "$releases" "$shipped"
  fi
  printf '</gno-frame>\n\n'
}

render_deployed() {
  local deployed="$STATE_DIR/deployed" date name path
  # A week with nothing new live keeps the old list: it ages out as the
  # ranking does, rather than sit under "0 deployments this week".
  is_recent deployed || return 0
  printf '<gno-icon name="box" /> **Recently deployed**\n\n'
  while IFS=$'\t' read -r date name path; do
    printf -- '- [%s](%s) · %s\n' "$name" "$path" "$(short_date "$date")"
  done <"$deployed"
}

render_used() {
  is_recent used || return 0
  printf '<gno-icon name="flame" /> **[Top realms](/r/)** · weekly users\n\n'
  awk -F'\t' '{ printf "%d. [%s](%s) · %d\n", NR, $2, $3, $1 }' "$STATE_DIR/used"
}

render_imported() {
  local imported="$STATE_DIR/imported"
  [ -s "$imported" ] || return 0
  printf '<gno-icon name="cube" /> **[Top packages](/p/)** · imports\n\n'
  awk -F'\t' '{ printf "%d. [%s](%s) · %d\n", NR, $2, $3, $1 }' "$imported"
}

render_releases() {
  local date url text
  [ -s "$STATE_DIR/release" ] || return 0
  printf '<gno-icon name="rocket" /> **[Gno.land releases](https://github.com/gnolang/gno/releases)**\n\n'
  awk -F'\t' 'NF == 3' "$STATE_DIR/release" | while IFS=$'\t' read -r date url text; do
    printf -- '- [%s](%s) · %s\n' "$text" "$url" "$(short_date "$date")"
  done
}

render_shipped() {
  local shipped="$STATE_DIR/shipped"
  [ -s "$shipped" ] || return 0
  printf '<gno-icon name="branch" /> **[Just merged on GitHub](https://github.com/gnolang/gno/pulls?q=is%%3Apr+is%%3Amerged)**\n\n'
  first_rows "${1:-$SHIPPED_SIZE}" <"$shipped" |
    awk -F'\t' '{ printf "- [%s](https://github.com/gnolang/gno/pull/%s)\n", $2, $1 }'
}

render_stat() {
  printf '<gno-frame>\n**%s** %s this week\n</gno-frame>' "$1" "$2"
}

render_builders() {
  local tag users realms packages recent ranking shipped
  recent=$(render_deployed)
  ranking=$(render_used)
  shipped=$(render_shipped)
  [ -n "$recent$ranking$shipped" ] || return 0

  printf '## <gno-icon name="signal" /> Live on %s\n\n' "$NETWORK"
  [ ! -s "$STATE_DIR/version" ] || {
    tag=$(cat "$STATE_DIR/version")
    printf 'Runs [%s](https://github.com/gnolang/gno/releases/tag/%s). ' "$tag" "$tag"
  }
  printf 'Watch it on [Gnockpit](https://gnockpit.%s.testnets.gno.land) or [Gnoscan](https://gnoscan.io/?chainId=%s-1).\n\n' "$NETWORK" "$NETWORK"
  if is_recent activity && is_recent deployed-count; then
    IFS=$'\t' read -r users realms <"$STATE_DIR/activity"
    packages=$(cat "$STATE_DIR/deployed-count")
    render_columns "$(render_stat "$packages" deployments)" "$(render_stat "$users" 'active users')" "$(render_stat "$realms" 'realms used')"
  fi
  render_columns "$recent" "$ranking" "$shipped"
  printf -- '---\n\n'
}

# Exactly one of each marker, in order: anything else would eat the page.
markers_ok() {
  [ -f "$1" ] && START="$2" END="$3" awk '
    $0 == ENVIRON["START"] { s++; if (e) bad = 1 }
    $0 == ENVIRON["END"] { e++ }
    END { exit !(s == 1 && e == 1 && !bad) }' "$1"
}

# A page may leave out the second section's markers, but not get them wrong:
# nested or interleaved, the first fill would overwrite one block with the other.
page_ok() {
  markers_ok "$1" "$START" "$END" || return 1
  ! grep -qF -e "$TRENDING_START" -e "$TRENDING_END" "$1" || {
    markers_ok "$1" "$TRENDING_START" "$TRENDING_END" &&
      { markers_ok "$1" "$END" "$TRENDING_START" || markers_ok "$1" "$TRENDING_END" "$START"; }
  }
}

# Reads a page and writes it with a block between two markers.
fill_block() {
  START="$1" END="$2" BLOCK="$3" awk '
    $0 == ENVIRON["START"] { print; printf "%s", ENVIRON["BLOCK"]; skip = 1; next }
    $0 == ENVIRON["END"] { skip = 0 }
    !skip'
}

# A failed reload stays pending, so that the next run tries it again.
reload() {
  [ -n "$ON_CHANGE" ] || return 0
  touch "$STATE_DIR/reload-pending"
  if sh -c "$ON_CHANGE"; then
    rm -f "$STATE_DIR/reload-pending"
  else
    echo "whatsnew: ON_CHANGE failed, the next run tries again" >&2
  fi
}

main() {
  local home="" dry_run=0 builders=0 render=render_block block trending="" tmp
  FORCE=0
  while [ $# -gt 0 ]; do
    case $1 in
    --dry-run) dry_run=1 ;;
    --force) FORCE=1 ;;
    --builders) builders=1 render=render_builders ;;
    -*)
      echo "whatsnew: unknown option $1" >&2
      exit 2
      ;;
    *) home=$1 ;;
    esac
    shift
  done
  if [ -z "$home" ] || ! page_ok "$home"; then
    echo "usage: whatsnew.sh [--dry-run] [--force] [--builders] <home.md>, with its two marker lines" >&2
    exit 2
  fi
  STATE_DIR=${STATE_DIR:-$(dirname "$home")/.whatsnew}
  export STATE_DIR
  mkdir -p "$STATE_DIR"

  refresh shipped "$FEED_TTL" &
  if [ "$builders" = 1 ]; then
    refresh version "$FEED_TTL" &
    [ -z "$INDEXER_URL" ] || refresh deployed "$USED_TTL" &
  else
    refresh x "$FEED_TTL" &
    refresh blog "$FEED_TTL" &
    refresh release "$FEED_TTL" &
    refresh imported "$IMPORTED_TTL" &
  fi
  [ -z "$INDEXER_URL" ] || refresh used "$USED_TTL" &
  wait

  block=$("$render"; printf x)
  block=${block%x}
  if [ "$builders" = 0 ] && grep -qF "$TRENDING_START" "$home"; then
    trending=$(render_trending; printf x)
    trending=${trending%x}
  fi
  # An outage on a host that lost its state must not wipe a filled block.
  if [ -z "$block$trending" ]; then
    echo "whatsnew: nothing to show, page left as it is" >&2
    return 0
  fi
  if [ "$dry_run" = 1 ]; then
    printf '%s%s' "$block" "$trending"
    return 0
  fi

  # Same folder and same mode, so the rename is atomic and gnoweb can read it.
  tmp=$(mktemp "$home.XXXXXX")
  PAGE_TMP=$tmp
  trap 'rm -f -- "$PAGE_TMP"' EXIT
  cp -p "$home" "$tmp"
  chmod u+w "$tmp"
  # cat reports a failed write, which busybox awk does not.
  if [ -n "$block" ]; then fill_block "$START" "$END" "$block"; else cat; fi <"$home" |
    if [ -n "$trending" ]; then fill_block "$TRENDING_START" "$TRENDING_END" "$trending"; else cat; fi |
    cat >"$tmp"
  # The mode bit, not -w: root can write any page, read-only or not.
  [ -n "$(find "$home" -perm -200)" ] || chmod u-w "$tmp"
  if cmp -s "$home" "$tmp"; then
    rm -f "$tmp"
    echo "whatsnew: unchanged"
    [ ! -f "$STATE_DIR/reload-pending" ] || reload
    return 0
  fi
  mv "$tmp" "$home"
  echo "whatsnew: updated $home"
  reload
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  case ${1:-} in
  --source) "source_$2" ;;
  --imports) abci "$RPC_URL" vm/qdoc "$2" | jq -r '.imports[]?' || echo '!failed' ;;
  *) main "$@" ;;
  esac
fi
