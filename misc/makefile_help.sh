#!/usr/bin/env bash
set -euo pipefail

# Initialize defaults
relative_to=""
wildcards=()
directories=()
makefile=""
print_help=false
expecting_wildcards=false
expecting_dirs=false

# Associative array to hold processed README banners
declare -A dir_descriptions
declare -A target_descriptions

function show_help {
    grep '^#>' "$0" | sed 's/^#>//'
}

#> Usage:
#>   makefile_help.sh <Makefile> [OPTIONS]
#>
#> Description:
#>   This script extracts and displays Makefile targets with inline comments,
#>   then processes a list of directories containing Makefiles, highlighting
#>   those with a `help` target by prefixing them with a `*`.
#>
#>   The Makefile argument is mandatory and must be a readable file.
#>
#>   Targets are extracted as lines of the form:
#>       target: prereqs   # description
#>
#>   These are sorted and presented with the associated descriptions, if any.
#>
#> Required:
#>   <Makefile>              Path to a readable Makefile
#>
#> Options:
#>   -r, --relative-to PATH  Treat PATH as the relative invocation path from the user's shell.
#>                           For example, if run as:
#>                               (cd bar/baz && script Makefile -r bar/baz)
#>                           then script assumes the user's real shell was at the grandparent
#>                           directory (e.g., /foo in /foo/bar/baz).
#>
#>   -d, --dirs DIR...       List of directories that may contain Makefiles to scan.
#>                           Each must be passed as a separate argument.
#>                           Optionally terminated by `--` to clarify end of list.
#>                           e.g.:
#>                               -d . contrib tests --
#>
#>   -w, --wildcard VALUE... Accepts multiple wildcard substitutions for % targets.
#>                           Optionally terminated by `--`, e.g.:
#>                               -w debug release test --
#>
#>   -h, --help              Display this help message and exit.
#>
#> Behavior:
#>   - Lists Makefile targets with descriptions from `#` comments.
#>   - Lists explicitly provided directories that contain Makefiles.
#>   - Prefixes entries with `*` if they include a `help` target.
#>   - If any wildcard or listed directory contains a `README.md`, its first line
#>     is extracted and appended (in parentheses) to the relevant output line.

# Process arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            print_help=true
            shift
            ;;
        -r|--relative-to)
            if [[ $# -lt 2 ]]; then
                echo "Error: --relative-to requires a path argument" >&2
                exit 1
            fi
            relative_to="$2"
            shift 2
            ;;
        -w|--wildcard)
            expecting_wildcards=true
            expecting_dirs=false
            shift
            ;;
        -d|--dirs)
            expecting_dirs=true
            expecting_wildcards=false
            shift
            ;;
        --)
            expecting_wildcards=false
            expecting_dirs=false
            shift
            ;;
        -*)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
        *)
            if $expecting_wildcards; then
                wildcards+=("$1")
                shift
            elif $expecting_dirs; then
                directories+=("$1")
                shift
            elif [[ -z "$makefile" ]]; then
                makefile="$1"
                shift
            else
                echo "Unexpected argument: $1" >&2
                exit 1
            fi
            ;;
    esac
done

# Help or missing Makefile
if $print_help; then
    show_help
    exit 0
elif [[ -z "$makefile" ]]; then
    echo "Error: A Makefile argument is required." >&2
    echo >&2
    show_help
    exit 1
elif [[ ! -f "$makefile" || ! -r "$makefile" ]]; then
    echo "Error: '$makefile' is not a readable Makefile." >&2
    exit 1
fi

mapfile -t targets < <(
  grep -E '^[A-Za-z][^:]*:' "$makefile" \
    | grep -v '#.*@LEGACY' \
    | sed -E 's/:.*//'
)

wildcard_filter="%"
wildcard_max_len=1
target_max_len=1
if [ "${#wildcards[@]}" -eq 0 ]; then
  # no filter
  wildcard_filter="."
  for ts in "${targets[@]}"; do
    (( ${#ts} > target_max_len )) && target_max_len=${#ts}
  done
else
  for ws in "${wildcards[@]}"; do
    (( ${#ws} > wildcard_max_len )) && wildcard_max_len=${#ws}
  done
  dots=$(printf '%*s' "$wildcard_max_len" '' | tr ' ' '.')
  for ts in "${targets[@]}"; do
    if [[ "$ts" == *%* ]]; then
      processed+=( "${ts//%/$dots}" )
    else
      processed+=( "$ts" )
    fi
  done
fi

merged_dirs=("${wildcards[@]}" "${directories[@]}")
IFS=$'\n' read -r -d '' -a all_dirs < <(
    printf '%s\n' "${merged_dirs[@]}" | sort -u && printf '\0'
)
for d in "${all_dirs[@]}"; do
    if [[ -d "$d" && -r "$d/README.md" ]]; then
        banner=$(head -n1 "$d/README.md" | sed -E \
            -e 's/^ *##* *//' \
            -e 's/^ *('"$d"'|`'"$d"'`) *((--*|:) *|$)//I' \
            -e 's/^(..*)$/ (\1)/'
        )
        dir_descriptions["$d"]="$banner"
    fi
done

# # Extract lines beginning with a-z then ':' excluding LEGACY tags
# function get_target_lines() {
#     grep -E '^[a-z][^:]*:' "$makefile" | grep -v '#.*@LEGACY'
# }

# # Debug print (placeholder for logic)
# echo "Makefile: $makefile"
# echo "Relative to: ${relative_to:-<none>}"
# echo "Wildcards: ${wildcards[*]:-<none>}"
# echo "Directories: ${directories[*]:-<none>}"



echo "Available make targets:"

( \
    cat "$makefile" | grep '^[a-z][^:]*:' | grep -v '#.*@LEGACY' | \
        grep -v "$wildcard_filter" | \
        sed \
            -e 's/:[^#]*# */# /' $\
    		-e 's/:[^#]*$$//' $\
    		-e 's/$/$(subst .,$(SPACE),$(call MAX_TARGET_CHARS,$(1)))   $(HASH)/' $\
    		-e 's/^\($(call MAX_TARGET_CHARS,$(1))...\) *$(HASH)/\1<--/' $\
    		-e 's/^/  /' \
        $(call SED_EXTRACT_TARGET_AND_COMMENT,$(1)) $(if $(1),; \
    for d in $(patsubst %/,%,$(1)) ; do \
        desc="$( \
        	head -1 "$d/README.md" 2> /dev/null | \
		        sed -E \
		            -e 's/^ *##* *//' \
		            -e 's/^ *('"$d"'|`'"$d"'`) *((--*|:) *|$)//I' \
		            -e 's/^(..*)$$/ (\1)/' || \
		    echo > /dev/null \
        )" ; \
        cat "$makefile" | grep '^[a-z][^:]*:' | grep -v '#.*@LEGACY' | \
            grep '\%' | \
            sed \
                -e 's,\%,'"$d,g" \
                $(call SED_EXTRACT_TARGET_AND_COMMENT,$(1)) \
                -e "s/\$/$desc/" ; \
    done,) \
) | \
sort
#	@$(call BASH_DISPLAY_TARGETS_AND_COMMENTS,$(programs))
#	@$(BASH_DISPLAY_SUB_MAKES)