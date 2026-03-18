#!/usr/bin/env bash
# Generate a GENESIS_TIME line to copy-paste into gen-genesis.sh.
#
# Uses the system `date` command (macOS BSD date and GNU date both supported).
set -eo pipefail

default_tz=$(date +%Z)

echo "Enter the desired genesis date and time."
echo ""

read -rp "Date and time (YYYY-MM-DD HH:MM): " input_datetime
read -rp "Timezone [${default_tz}]: " input_tz
input_tz="${input_tz:-$default_tz}"

# Convert to Unix timestamp — macOS (BSD) and Linux (GNU) use different flags.
if date --version &>/dev/null 2>&1; then
  # GNU date (Linux)
  ts=$(TZ="$input_tz" date -d "$input_datetime" +%s)
else
  # BSD date (macOS) — append :00 seconds, otherwise date fills in the current clock seconds.
  ts=$(TZ="$input_tz" date -jf "%Y-%m-%d %H:%M:%S" "${input_datetime}:00" +%s)
fi

# Build the human-readable comment using date for formatting.
day=$(TZ="$input_tz" date -jf "%s" "$ts" "+%-d" 2>/dev/null || TZ="$input_tz" date -d "@$ts" "+%-d")
case $((day % 10)) in
  1) [[ "$day" != 11 ]] && suffix="st" || suffix="th" ;;
  2) [[ "$day" != 12 ]] && suffix="nd" || suffix="th" ;;
  3) [[ "$day" != 13 ]] && suffix="rd" || suffix="th" ;;
  *) suffix="th" ;;
esac

comment=$(TZ="$input_tz" date -jf "%s" "$ts" "+%A, %B $day$suffix %Y %H:%M GMT%z (%Z)" 2>/dev/null \
       || TZ="$input_tz" date -d "@$ts" "+%A, %B $day$suffix %Y %H:%M GMT%z (%Z)")

echo ""
echo "GENESIS_TIME=$ts # $comment"
