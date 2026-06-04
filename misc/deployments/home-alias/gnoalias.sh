#!/bin/sh

# Usage: ./gnoalias <alias_folder> <reload_cmd> <rpc_url>

# time to sleep between attempts
SLEEPING_TIME=60
# home folder in this context
ALIAS_HOME_FOLDER="${1:-./home}"
# target rpc node url
RPC_NODE_URL="${3:-https://rpc.gno.land}"
# docker restart gnoweb
DEFAULT_RELOAD_CMD="docker ps -a --filter ancestor=\"ghcr.io/gnolang/gno/gnoweb:master\" --format \"{{.ID}}\" | \
  xargs -r docker restart"
# Command executed when reload is needed
if [ -n "$2" ]; then
  RELOAD_CMD="$2"
else
  RELOAD_CMD="$DEFAULT_RELOAD_CMD"
fi

check_gnoland_home() {
  # Enters folder
  cd $ALIAS_HOME_FOLDER

  # if this file exists bypass the script homepage manually.
  if [ -f home-override.md ]; then
    echo "Home page overriden by home-override.md"
    if cmp -s home-override.md home.md; then # overriden not needed but what to stop script here anyway
      return 1
    else
      cp home-override.md home.md
      return 0
    fi
  fi

  # tries to query r/gnoland/home.
  echo "Querying current home..."
  if gnokey query vm/qrender -remote ${RPC_NODE_URL} -data "gno.land/r/gnoland/home:" > out.md; then
    # eventually patch the page to add things like the newsletter html form
   [ -f extra-blocks.md ] && cat extra-blocks.md >> out.md
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
    echo "Restarting Gnoweb service deliberately"
    eval "$RELOAD_CMD"
  fi
  echo "Sleeping for ${SLEEPING_TIME} seconds..."
  sleep ${SLEEPING_TIME} # wait next round
done
