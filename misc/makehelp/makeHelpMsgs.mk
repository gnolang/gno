
## @var $(DIR_OFFSET_OPT)
## @brief Attempts to infer the argument passed to `make -C`, if any.
##
## Uses a heuristic to (unreliably) detect whether `make` was invoked with
## the `-C` option by comparing the parent shell's current working directory
## (`PWD`) to the directory `make` believes it is running in (`CURDIR`).
##
## If a difference is found, this macro computes a relative path from `PWD`
## to `CURDIR`, which approximates the argument passed to `-C`.
## If no difference is detected, it yields an empty string.
##
## @warning
##   This inference is not guaranteed to be correct in all cases and may
##   produce incorrect results.
##
## @return
##   A string suitable for passing to the `-r` option of the helper tool,
##   or empty if no offset is detected.
DIR_OFFSET_OPT    = $(if $(filter $(PWD),$(CURDIR)),,$(patsubst $(patsubst %/,%,$(PWD))/%,%,$(CURDIR)))

## @fn $(call RUN_MAKEFILE_HELP,repo_root_relpath,wildcard_values)
## @brief Invoke the `makefile_help.go` CLI helper with proper flags.
##
## @param repo_root_relpath
##   The relative path  from the current working directory to the git
##   repository root. This prefix is used to locate `misc/makehelp/makefile_help.go`.
##
## @param wildcard_values
##   A space-separated list of values to substitute for `%` targets
##   when expanding wildcard rules.
##
## @details
##   1. Runs the helper via `go run $(1)/misc/makehelp/makefile_help.go`.
##   2. Adds `-r DIR_OFFSET_OPT` if `DIR_OFFSET_OPT` is non-empty.
##   3. Scans all subdirectories for `Makefile` and passes each one with `-d`.
##   4. For each wildcard value, adds a `-w "VALUE"` flag.
##   5. Targets the `Makefile` from the current directory to produce formatted
##          help output.
##
## @example
##   # Run helper from a subdir, relative to the repo root, with wildcards:
##   $(call RUN_MAKEFILE_HELP, ../, foo bar)
RUN_MAKEFILE_HELP = \
    go run $(if $(filter-out . ./,$(1)),$(patsubst %/,%,$(1))/,)misc/makehelp/makefile_help.go \
        $(if $(DIR_OFFSET_OPT),-relative-to "$(DIR_OFFSET_OPT)",) \
        $(foreach makeDir,$(patsubst %/Makefile,%,$(wildcard */Makefile)),-dir "$(makeDir)") \
        $(foreach wildCardValue,$(2),-wildcard "$(wildCardValue)") \
        Makefile
