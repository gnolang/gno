#!/bin/bash
# the stupid build tool.
# to use this in your projects, see: https://zxq.co/rosa/tool
set -eo pipefail

# provides documentation for the sub-commands.
DOCUMENTATION=(
	'help [COMMAND]'
	"returns help for the specified command, or for all commands if none is specified.
		$ ./tool help
		$ ./tool help install"

	'install [<FLAGS>...] [<PROGRAM>...]'
	"installs the given program(s). available programs:
			- gnokey:    gno keypair management tool
			- gno:       gno development toolkit
			- gnodev:    hot-reloading gnoweb + gnoland for realm development
			- gnoland:   gno.land blockchain node
			- gnoweb:    web interface to the blockchain node
			- gnofaucet: faucet API and interface
		EXPERIMENTAL:
			- gnomd:     markdown terminal renderer
			- gnokeykc:  gnokey with system keychain support
		$ ./tool install gnokey            # install a single program
		$ ./tool install gnokey gno gnodev # for a simple development environment"

	'fmt [<FLAGS>...] [<PATH>...]'
	"runs gofumpt on the given directories. available flags:
			-version  show version and exit

			-d        display diffs instead of rewriting files
			-e        report all errors (not just the first 10 on different lines)
			-l        list files whose formatting differs from gofumpt's
			-w        write result to (source) file instead of stdout
			-extra    enable extra rules which should be vetted by a human

			-lang       str    target Go version in the form \"1.X\" (default from go.mod)
			-modpath    str    Go module path containing the source file (default from go.mod)
		with no flags, runs with \"-w .\" by default.
		$ ./tool fmt          # format current directory and subdirectories.
		$ ./tool fmt -l ./... # check formatting in current directory and subdirectories."

	'lint [<ARGS>...]'
	"lint the project using golangci-lint. with no flag, lints the whole project
		with the config used on the GitHub CI.
		$ ./tool lint              # lint full project
		$ ./tool lint cache status # using a golangci-lint subcommand"

	'tidy [<FLAGS>...] [<DIRECTORY>...]'
	"run \`go mod tidy\` on the given directories.
		if directories are not given, runs tidy on all directories containing a
		go.mod file.
		$ ./tool tidy
		$ ./tool tidy -v ./contribs/gnodev"

	'tool-anywhere-source'
	"prints the source of tool-anywhere. tool-anywhere allows you to call tool
		from any subdirectory.
		# set up your ~/bin directory if you haven't already
		$ mkdir -p ~/bin                           # create ~/bin if it doesn't exist
		$ echo 'export PATH=\"\$PATH:\$HOME/bin\"' >> ~/.profile
		$ source ~/.profile                        # re-load profile
		# add tool-anywhere!
		$ ./tool tool-anywhere-source > ~/bin/tool # save to your ~/bin directory
		$ chmod +x ~/bin/tool                      # mark as executable
		$ cd gnovm; tool help                      # success!"

	# TODO: tool run <program> - runs the given program using go run
	# (ie. with up-to-date codebase), ie. `tool run gno test -v ./...`
)

# absolute path to script
tool_root="$(dirname "$(realpath "$0")")"

# list of programs available for `tool install`
# to add a new program, add it to PROGRAMS, the DOCUMENTATION, and create
# PROGRAMS_$name.
PROGRAMS=(
	'gnokey'
	'gno'
	'gnodev'
	'gnoland'
	'gnoweb'
	'gnofaucet'
	'gnomd'
	'gnokeykc'
)
PROGRAMS_gnokey=(
	# directory of go.mod
	'.'
	# location (relative to go.mod dir)
	'./gno.land/cmd/gnokey'
)
PROGRAMS_gno=(
	'.'
	'./gnovm/cmd/gno'
	# extra go build/install flags
	"-ldflags -X=github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$tool_root"
)
PROGRAMS_gnodev=(
	'./contribs/gnodev'
	'./cmd/gnodev'
	"-ldflags -X=github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$tool_root"
)
PROGRAMS_gnoland=(
	'.'
	'./gno.land/cmd/gnoland'
)
PROGRAMS_gnoweb=(
	'.'
	'./gno.land/cmd/gnoweb'
)
PROGRAMS_gnofaucet=(
	'./contribs/gnofaucet'
	'.'
)
PROGRAMS_gnomd=(
	'./contribs/gnomd'
	'.'
)
PROGRAMS_gnokeykc=(
	'./contribs/gnokeykc'
	'.'
)

green() {
	printf "\033[0;32m%s\033[0m\n" "$1"
}

tool() {
	cmd="$1"
	if [ $# -ne 0 ]; then
		shift
	fi
	case "$cmd" in
		install)
			local flags=""
			for program_name in "$@"; do
				if [[ $program_name == -* ]]; then
					# TODO: not perfect; does not support flag arguments, unless using -flag= syntax.
					flags="$flags $program_name"
					continue
				fi
				if ! printf '%s\0' "${PROGRAMS[@]}" | grep -F -x -z -- "$program_name" > /dev/null; then
					echo "program $program_name does not exist; \`./tool help install\` for a list of available programs"
					continue
				fi

				local program_chdir=PROGRAMS_$program_name[0]
				local program_path=PROGRAMS_$program_name[1]
				local program_build=PROGRAMS_$program_name[2]
				(
					# prints command as it executes
					set -o xtrace
					go install -C "$tool_root/"${!program_chdir} ${!program_build} ${flags} ${!program_path}
				)
				green "+ $program_name installed"
			done
			;;

		fmt)
			if [ $# -eq 0 ]; then
				tool rundep mvdan.cc/gofumpt -w .
				# TODO: format gnovm/stdlibs and examples gno files.
				return 0
			fi
			tool rundep mvdan.cc/gofumpt "$@"
			# TODO: evaluate `goimports` support
			;;

		lint)
			if [ $# -eq 0 ]; then
				# cd to root, so we can lint entire project.
				cd "$tool_root"
				tool rundep github.com/golangci/golangci-lint/cmd/golangci-lint run --config "$tool_root/".github/golangci.yml
				return 0
			fi
			tool rundep github.com/golangci/golangci-lint/cmd/golangci-lint "$@"
			;;

		tidy)
			local flags=""
			local executed=0
			for directory in "$@"; do
				if [[ $directory == -* ]]; then
					# TODO: not perfect; does not support flag arguments, unless using -flag= syntax.
					flags="$flags $directory"
					continue
				fi
				executed=1
				(
					# prints command as it executes them
					set -o xtrace
					env -C $directory go mod tidy $flags
				)
			done
			if [ $executed -eq 0 ]; then
				(
					# prints command as it executes
					set -o xtrace
					find . -name "go.mod" -execdir go mod tidy $flags \;
				)
			fi
			;;

		tool-anywhere-source)
			cat << 'EOF'
#!/bin/bash
# tool anywhere!
# install and +x in your $PATH
pwd="$PWD"
while [ ! -f './tool' -a "$PWD" != '/' ]; do
	cd ..
done
if [ "$PWD" = '/' ]; then
	echo 'tool not found'
	exit 127
fi
env -C "$pwd" "$(realpath ./tool)" "$@"
exit $?
EOF
			;;


		# ----------------------------------------------------------------------
		# INTERNAL
		rundep)
			(
				set -o xtrace
				go run -v -modfile "$tool_root/"misc/devdeps/go.mod "$@"
			)
			;;

		# ----------------------------------------------------------------------
		*)
			BOLD='\033[1m'
			NC='\033[0m'

			# "help" explicitly called - print program line.
			if [ "$cmd" == 'help' ]; then
				printf "${BOLD}tool${NC} - the stupid build tool\n"
				echo
			fi
			printf "Usage: ${BOLD}./tool <SUBCOMMAND> [<ARGUMENT>...]${NC}\n"
			if [ -z "$1" ]; then
				echo 'Subcommands:'
			fi

			idx=0
			found=false
			while [ "${DOCUMENTATION[idx]}" ]; do
				stringarr=(${DOCUMENTATION[idx]})
				# if $1 is set and this is the wrong command, then skip it
				if [ -n "$1" -a "${stringarr[0]}" != "$1" ]; then
					((idx+=2))
					continue
				fi

				found=true

				echo
				printf "\t${BOLD}${DOCUMENTATION[idx]}${NC}\n"
				printf "\t\t${DOCUMENTATION[idx+1]}\n"

				((idx+=2))
			done
			if [ "$found" = false ]; then
				echo
				echo "The specified subcommand $1 could not be found."
			fi
			;;
	esac
}

tool "$@"
