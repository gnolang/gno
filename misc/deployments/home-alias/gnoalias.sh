#!/bin/sh

# Usage: ./gnoalias <alias_folder> [reload_cmd] [rpc_url]
#
# gnoweb re-reads a static alias file when it changes, so writing home.md is
# enough. reload_cmd is only for a gnoweb older than that.
#
# BLOG_LINK_BASE (optional): prefix for the blog post links, https://gno.land on
# a testnet. The posts always come from mainnet: the testnet blogs are empty.

SLEEPING_TIME=3600
# home folder in this context
ALIAS_HOME_FOLDER="${1:-./home}"
# target rpc node url
RPC_NODE_URL="${3:-https://rpc.gno.land}"
# line of home-override.md replaced by the latest blog posts
BLOG_MARKER="<!-- latest-blogposts -->"
RELOAD_CMD="$2"

# qrender <rpc> <pkgpath:path> prints the realm's markdown, without the
# "height:" line and "data: " prefix gnokey adds.
qrender() {
  out=$(gnokey query vm/qrender -remote "$1" -data "$2") || return 1
  printf '%s\n' "$out" | awk '!d && /^height: / { next } !d && sub(/^data: /, "") { d = 1 } { print }'
}

# latest_blogposts prints the first row of the blog index, as laid out by
# RenderHome in p/gnoland/blog/v0: the three most recent posts in a
# <gno-columns> grid. Titles become bold paragraphs above their date, so they
# stay out of the page's table of contents.
latest_blogposts() {
  qrender https://rpc.gno.land "gno.land/r/gnoland/blog:" | awk -v base="$BLOG_LINK_BASE" '
    /^<gno-columns>$/ { on = 1 }
    !on { next }
    { gsub(/\]\(\/r\/gnoland\/blog/, "](" base "/r/gnoland/blog") }
    /^### / { print "**" substr($0, 5) "**"; print ""; next }
    { print }
    /^<\/gno-columns>$/ { exit }'
}

# render_override prints home-override.md with the blog marker replaced.
render_override() {
  posts=
  if grep -qxF "$BLOG_MARKER" home-override.md; then
    posts=$(latest_blogposts)
    [ -n "$posts" ] || return 1
  fi
  POSTS=$posts awk -v marker="$BLOG_MARKER" \
    '$0 == marker { print ENVIRON["POSTS"]; next } { print }' home-override.md
}

check_gnoland_home() {
  # Enters folder
  cd $ALIAS_HOME_FOLDER

  # if this file exists bypass the script homepage manually.
  if [ -f home-override.md ]; then
    echo "Home page overriden by home-override.md"
    # a failed blog query keeps the current page rather than dropping the posts
    if ! render_override > out.md; then
      echo "Warn: failed to read the latest blog posts."
      rm -f out.md
      [ -f home.md ] && return 1
      # first round: gnoweb cannot start without home.md, so serve the page
      # without the posts until the next round
      grep -vxF "$BLOG_MARKER" home-override.md > home.md
      return 0
    fi
    if cmp -s out.md home.md; then # nothing changed since the last round
      rm out.md
      return 1
    fi
    mv out.md home.md
    return 0
  fi

  # tries to query r/gnoland/home.
  echo "Querying current home..."
  if qrender "${RPC_NODE_URL}" "gno.land/r/gnoland/home:" > out.md; then
    # eventually patch the page to add things like the newsletter html form
   [ -f extra-blocks.md ] && cat extra-blocks.md >> out.md
    if cmp -s out.md home.md; then # nothing changed since the last round
      rm out.md
      return 1
    fi
    mv out.md home.md
    echo "Success: gno.land home gathered"
  else
    echo "Warn: failed to update."
    return 1
  fi

  return 0
}

while true; do
  echo "Checking home"
  if check_gnoland_home; then
    echo "Home updated"
    [ -n "$RELOAD_CMD" ] && eval "$RELOAD_CMD"
  fi
  echo "Sleeping for ${SLEEPING_TIME} seconds..."
  sleep ${SLEEPING_TIME} # wait next round
done
