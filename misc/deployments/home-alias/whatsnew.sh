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
# Without an indexer, "Top realms" stays out of the page.
: "${INDEXER_URL:=}"
: "${X_ACCOUNT:=_gnoland}"
: "${YOUTUBE_CHANNEL:=UCHlHaup7HEDQFQrb1goxAkg}"
# Videos whose title matches this regular expression stay out of the feed.
: "${VIDEO_SKIP=Jae Kwon}"
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
export RPC_URL BLOG_RPC_URL LINK_BASE INDEXER_URL X_ACCOUNT YOUTUBE_CHANNEL VIDEO_SKIP DEPLOYED_SKIP USED_NAMESPACES

START='<!-- whats-new:start -->'
END='<!-- whats-new:end -->'
X_SLOTS=3
OTHER_SLOTS=2
# An old blog post, video or release would not read as new.
MAX_ROW_AGE=60
LIST_SIZE=6
SHIPPED_SIZE=3
MIN_CHARS=20
MAX_IMPORT_FAILURES=10
MAX_CHARS=110
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

# X rewrites a bare gno.land into a t.co link; every other link is dropped, so
# that the page never carries a link a post chose.
expand_links() {
  local text url target
  text=$(cat)
  for url in $(grep -oE 'https://t\.co/[A-Za-z0-9]+' <<<"$text" | sort -u || true); do
    target=$(curl -s --max-time 10 -o /dev/null -w '%{redirect_url}' "$url" || true)
    if grep -qE '^https?://(www\.)?gno\.land(/|$)' <<<"$target"; then
      text=${text//$url/Gno.land}
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

# A source prints one row per item: date (YYYYMMDD), kind, link, escaped text.
row() {
  [[ $1 =~ ^[0-9]{8}$ ]] || return 1
  printf '%s\t%s\t%s\t%s\n' "$1" "$2" "$3" "$(to_md <<<"$4")"
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
      | [.id, (if .article then "article" else "x" end)] | @tsv')
  while IFS=$'\t' read -r id kind; do
    [ "$n" -lt 6 ] || break
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
    row "$date" x "$url" "$text"
    n=$((n + 1))
  done <<<"$posts"
  [ "$n" -gt 0 ]
}

source_blog() {
  local fields date slug title
  fields=$(abci "$BLOG_RPC_URL" vm/qrender 'gno.land/r/gnoland/blog:' |
    jq -Rrs --arg months "$MONTHS" "$YMD"'
      first(capture("### \\[(?<t>[^\\n]*)\\]\\(/r/gnoland/blog:p/(?<s>[a-z0-9-]+)\\)\\n(?<d>[0-9]{2}) (?<m>[A-Z][a-z]{2}) (?<y>[0-9]{4})"; "g"))
      | [ymd, .s, .t] | @tsv')
  [ -n "$fields" ]
  IFS=$'\t' read -r date slug title <<<"$fields"
  row "$date" blog "${LINK_BASE}/r/gnoland/blog:p/${slug}" "$title"
}

source_video() {
  local fields id date title
  fields=$("${CURL[@]}" "https://www.youtube.com/feeds/videos.xml?channel_id=${YOUTUBE_CHANNEL}" |
    jq -Rrs --arg skip "$VIDEO_SKIP" '
      gsub("\\s+"; " ")
      | first(capture("<entry>.*?<yt:videoId>(?<id>[A-Za-z0-9_-]{11})</yt:videoId>.*?<title>(?<t>[^<]*)</title>.*?<published>(?<y>[0-9]{4})-(?<m>[0-9]{2})-(?<d>[0-9]{2})"; "g")
          | select($skip == "" or (.t | test($skip; "i") | not)))
      | [.id, "\(.y)\(.m)\(.d)", .t] | @tsv')
  [ -n "$fields" ]
  IFS=$'\t' read -r id date title <<<"$fields"
  row "$date" video "https://www.youtube.com/watch?v=${id}" "$(html_to_text <<<"$title" | expand_links)"
}

source_release() {
  local fields tag date
  fields=$("${CURL[@]}" https://api.github.com/repos/gnolang/gno/releases/latest |
    jq -er 'select(.tag_name | test("^v[0-9.]+$")) | [.tag_name, (.published_at[0:10] | gsub("-"; ""))] | @tsv')
  IFS=$'\t' read -r tag date <<<"$fields"
  row "$date" release "https://github.com/gnolang/gno/releases/tag/${tag}" "Gno.land ${tag} released"
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
    jq -r ".data.getTransactions | if type == \"array\" then .[] else error(\"no data\") end
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
  probe=$(block_time $((tip - RATE_PROBE))) || return 1
  lower=$((tip - RATE_PROBE * ACTIVITY_WINDOW / (head - probe)))
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

# One line per project, newest first; the week's count goes beside it.
source_deployed() {
  local rows height name path epoch
  rows=$(scan_week 'MsgAddPackage: { package: { path: { like: "." } } }' '... on MsgAddPackage { package { path } }' \
    'select(.__typename == "MsgAddPackage") | [$height, .package.path] | @tsv' |
    grep -E $'^[0-9]+\tgno\\.land/[rp]/[a-z0-9_./-]+$' | grep -vF .. |
    awk -v skip="$DEPLOYED_SKIP" 'skip == "" || $0 !~ skip' | sort -rn)
  grep -c . <<<"$rows" >"$STATE_DIR/deployed-count"
  awk -F'\t' '{ split($2, p, "/"); if (!((p[3] "/" p[4]) in seen)) { seen[p[3] "/" p[4]]; print } }' <<<"$rows" |
    first_rows "$LIST_SIZE" | short_paths | while IFS=$'\t' read -r height name path; do
    epoch=$(block_time "$height")
    printf '%s\t%s\t%s\n' "$(jq -nr --argjson t "$epoch" '$t | strftime("%Y%m%d")')" "$(cut -d/ -f1,2 <<<"$name")" "$path"
  done
}

# A testnet runs each release first, as a candidate.
source_candidate() {
  "${CURL[@]}" 'https://api.github.com/repos/gnolang/gno/releases?per_page=10' |
    jq -er 'first(.[].tag_name | select(test("^v[0-9.]+-rc\\.[0-9]+$")))'
}

source_shipped() {
  local prs number title n=0
  prs=$("${CURL[@]}" 'https://api.github.com/search/issues?q=repo:gnolang/gno+is:pr+is:merged&sort=updated&order=desc&per_page=30' |
    jq -r --argjson size "$SHIPPED_SIZE" '[.items[] | select(.title | test("^(feat|fix)[(:!]"))] | sort_by(.closed_at) | reverse | .[:$size][]
      | (.title | sub("^[a-z]+(\\([^)]*\\))?!?: *"; "")) as $t
      | [.number, ($t[:1] | ascii_upcase) + $t[1:]] | @tsv')
  [ -n "$prs" ]
  while IFS=$'\t' read -r number title; do
    printf '%s\t%s\n' "$number" "$(MAX_CHARS=60 to_md <<<"$title")"
    n=$((n + 1))
  done <<<"$prs"
  [ "$n" -gt 0 ]
}

source_imported() {
  local imports
  imports=$({ abci "$RPC_URL" vm/qpaths gno.land/p/; abci "$RPC_URL" vm/qpaths gno.land/r/; } |
    grep . | xargs -P 4 -I{} "$BASH" "$0" --imports {})
  # A package may fail to read; many failing would skew the ranking.
  [ "$(grep -c '^!failed$' <<<"$imports" || true)" -le "$MAX_IMPORT_FAILURES" ]
  grep -E '^gno\.land/p/[a-z0-9_./-]+$' <<<"$imports" | grep -vF .. |
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

fragments() {
  local name
  for name in "$@"; do
    [ ! -f "$STATE_DIR/$name" ] || cat "$STATE_DIR/$name"
  done
}

newest() {
  sort -rs -t$'\t' -k1,1
}

# Laid out like the page's cards: a short green link, then neutral text.
render_rows() {
  local date kind url text icon label first=1
  newest | while IFS=$'\t' read -r date kind url text; do
    case $kind in
    x) icon='𝕏' label="Gno.land \\@${X_ACCOUNT/_/\\_}" ;;
    blog) icon='<gno-icon name="book" />' label=Blog ;;
    video) icon='<gno-icon name="camera" />' label=Video ;;
    release) icon='<gno-icon name="rocket" />' label=Release ;;
    esac
    [ "$first" = 1 ] || printf '\n---\n\n'
    first=0
    printf '%s **[%s · %s](%s)**\n\n%s\n' "$icon" "$label" "$(short_date "$date")" "$url" "$text"
  done
}

render_posts() {
  fragments x | newest | first_rows "$X_SLOTS" | render_rows
}

render_news() {
  local cutoff rows shipped="$STATE_DIR/shipped"
  cutoff=$(jq -nr --argjson days "$MAX_ROW_AGE" 'now - $days * 86400 | strftime("%Y%m%d")')
  rows=$(fragments blog video release | awk -F'\t' -v cutoff="$cutoff" '$1 >= cutoff' |
    newest | first_rows "$OTHER_SLOTS" | render_rows)
  [ -z "$rows" ] || printf '%s\n' "$rows"
  [ -s "$shipped" ] || return 0
  [ -z "$rows" ] || printf '\n---\n\n'
  render_shipped
}

render_list() {
  awk -F'\t' '{ printf "%d. [%s](%s) · %d\n", NR, $2, $3, $1 }' "$1"
}

render_panel() {
  local used="$STATE_DIR/used" imported="$STATE_DIR/imported"
  if [ -s "$used" ] && ! older_than "$used" "$USED_MAX_AGE"; then
    printf '<gno-icon name="flame" /> **[Top realms](/r/)** · weekly users\n\n'
    render_list "$used"
    printf '\n'
  fi
  if [ -s "$imported" ]; then
    printf '<gno-icon name="cube" /> **[Top packages](/p/)** · imports\n\n'
    render_list "$imported"
  fi
}

# Three columns, like the rest of the page: the rows line up with the sections
# above and below.
render_block() {
  local posts news panel column first=1
  posts=$(render_posts)
  news=$(render_news)
  panel=$(render_panel)
  [ -n "$posts$news" ] || return 0
  printf '## <gno-icon name="newspaper" /> What'\''s new\n\nThe latest from Gno.land, and what people use on it.\n\n<gno-columns>\n'
  for column in "$posts" "$news" "$panel"; do
    [ -n "$column" ] || continue
    [ "$first" = 1 ] || printf '\n<gno-columns-sep>\n'
    printf '\n%s\n' "$column"
    first=0
  done
  printf '\n</gno-columns>\n\n'
  printf '<gno-button href="https://x.com/%s" label="Follow Gno.land on X" variant="outline" /> <gno-button href="%s/blog" label="Read the blog" variant="outline" />\n\n---\n\n' "$X_ACCOUNT" "$LINK_BASE"
}

render_shipped() {
  local shipped="$STATE_DIR/shipped"
  [ -s "$shipped" ] || return 0
  printf '<gno-icon name="branch" /> **[Just shipped on GitHub](https://github.com/gnolang/gno/pulls?q=is%%3Apr+is%%3Amerged)**\n\n'
  awk -F'\t' '{ printf "- [%s](https://github.com/gnolang/gno/pull/%s)\n", $2, $1 }' "$shipped"
}

render_builders() {
  local deployed="$STATE_DIR/deployed" used="$STATE_DIR/used" tag date name path users realms packages first=1 column
  local recent="" ranking="" shipped
  if [ -s "$deployed" ]; then
    recent='<gno-icon name="box" /> **Recently deployed**'$'\n\n'
    while IFS=$'\t' read -r date name path; do
      recent+="- [${name}](${path}) · $(short_date "$date")"$'\n'
    done <"$deployed"
  fi
  if [ -s "$used" ] && ! older_than "$used" "$USED_MAX_AGE"; then
    ranking=$(printf '<gno-icon name="flame" /> **[Top realms](/r/)** · weekly users\n\n'; render_list "$used")
  fi
  shipped=$(render_shipped)
  [ -n "$recent$ranking$shipped" ] || return 0

  printf '## <gno-icon name="signal" /> Live on %s\n\n' "$NETWORK"
  [ ! -s "$STATE_DIR/candidate" ] || {
    tag=$(cat "$STATE_DIR/candidate")
    printf 'Latest candidate: [%s](https://github.com/gnolang/gno/releases/tag/%s). ' "$tag" "$tag"
  }
  printf 'Watch it on [Gnockpit](https://gnockpit.%s.testnets.gno.land) or [Gnoscan](https://gnoscan.io/?chainId=%s-1).\n\n' "$NETWORK" "$NETWORK"
  if [ -s "$STATE_DIR/activity" ] && [ -s "$STATE_DIR/deployed-count" ] &&
    ! older_than "$STATE_DIR/activity" "$USED_MAX_AGE" && ! older_than "$STATE_DIR/deployed-count" "$USED_MAX_AGE"; then
    IFS=$'\t' read -r users realms <"$STATE_DIR/activity"
    packages=$(cat "$STATE_DIR/deployed-count")
    printf '<gno-columns>\n<gno-frame>\n**%s** deployments this week\n</gno-frame>\n<gno-columns-sep>\n<gno-frame>\n**%s** active users this week\n</gno-frame>\n<gno-columns-sep>\n<gno-frame>\n**%s** realms used this week\n</gno-frame>\n</gno-columns>\n\n' \
      "$packages" "$users" "$realms"
  fi
  printf '<gno-columns>\n'
  for column in "$recent" "$ranking" "$shipped"; do
    [ -n "$column" ] || continue
    [ "$first" = 1 ] || printf '\n<gno-columns-sep>\n'
    printf '\n%s\n' "$column"
    first=0
  done
  printf '\n</gno-columns>\n\n---\n\n'
}

# Exactly one of each marker, in order: anything else would eat the page.
markers_ok() {
  [ -f "$1" ] && START="$START" END="$END" awk '
    $0 == ENVIRON["START"] { s++; if (e) bad = 1 }
    $0 == ENVIRON["END"] { e++ }
    END { exit !(s == 1 && e == 1 && !bad) }' "$1"
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
  local home="" dry_run=0 builders=0 block tmp
  FORCE=0
  while [ $# -gt 0 ]; do
    case $1 in
    --dry-run) dry_run=1 ;;
    --force) FORCE=1 ;;
    --builders) builders=1 ;;
    *) home=$1 ;;
    esac
    shift
  done
  if [ -z "$home" ] || ! markers_ok "$home"; then
    echo "usage: whatsnew.sh [--dry-run] [--force] [--builders] <home.md>, with its two marker lines" >&2
    exit 2
  fi
  STATE_DIR=${STATE_DIR:-$(dirname "$home")/.whatsnew}
  export STATE_DIR
  mkdir -p "$STATE_DIR"

  refresh shipped "$FEED_TTL" &
  if [ "$builders" = 1 ]; then
    refresh candidate "$FEED_TTL" &
    [ -z "$INDEXER_URL" ] || refresh deployed "$USED_TTL" &
  else
    refresh x "$FEED_TTL" &
    refresh blog "$FEED_TTL" &
    refresh video "$FEED_TTL" &
    refresh release "$FEED_TTL" &
    refresh imported "$IMPORTED_TTL" &
  fi
  [ -z "$INDEXER_URL" ] || refresh used "$USED_TTL" &
  wait

  if [ "$builders" = 1 ]; then
    block=$(render_builders; printf x)
  else
    block=$(render_block; printf x)
  fi
  block=${block%x}
  # An outage on a host that lost its state must not wipe a filled block.
  if [ -z "$block" ]; then
    echo "whatsnew: nothing to show, page left as it is" >&2
    return 0
  fi
  if [ "$dry_run" = 1 ]; then
    printf '%s' "$block"
    return 0
  fi

  # Same folder and same mode, so the rename is atomic and gnoweb can read it.
  tmp=$(mktemp "$home.XXXXXX")
  cp -p "$home" "$tmp"
  START="$START" END="$END" BLOCK="$block" awk '
    $0 == ENVIRON["START"] { print; printf "%s", ENVIRON["BLOCK"]; skip = 1; next }
    $0 == ENVIRON["END"] { skip = 0 }
    !skip' "$home" >"$tmp"
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
