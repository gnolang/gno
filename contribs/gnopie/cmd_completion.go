package main

import (
	"context"
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/commands"
)

func newCompletionCmd(io commands.IO) *commands.Command {
	return commands.NewCommand(
		commands.Metadata{
			Name:       "completion",
			ShortUsage: "gnopie completion <bash|zsh|fish>",
			ShortHelp:  "Generate shell completion scripts.",
			LongHelp: `Generate shell completion scripts for gnopie.

Usage:
  eval "$(gnopie completion bash)"
  gnopie completion zsh > ~/.zsh/completions/_gnopie
  gnopie completion fish > ~/.config/fish/completions/gnopie.fish`,
		},
		nil,
		func(_ context.Context, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: gnopie completion <bash|zsh|fish>")
			}
			switch args[0] {
			case "bash":
				io.Println(bashCompletion)
			case "zsh":
				io.Println(zshCompletion)
			case "fish":
				io.Println(fishCompletion)
			default:
				return fmt.Errorf("unsupported shell %q (use bash, zsh, or fish)", args[0])
			}
			return nil
		},
	)
}

const bashCompletion = `_gnopie() {
    local cur prev words cword
    _init_completion || return

    local verbs="GET EVAL READ INSPECT CALL RUN"
    local subcmds="config completion version"
    local flags="--home --key --json --quiet --send --gas-wanted --gas-fee --dry-run --generate-gnokey"

    case "$prev" in
        gnopie)
            COMPREPLY=( $(compgen -W "$verbs $subcmds $flags" -- "$cur") )
            return ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            return ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "$flags" -- "$cur") )
    fi
}
complete -F _gnopie gnopie`

const zshCompletion = `#compdef gnopie

_gnopie() {
    local -a verbs subcmds flags
    verbs=(GET EVAL READ INSPECT CALL RUN)
    subcmds=(config completion version)
    flags=(--home --key --json --quiet --send --gas-wanted --gas-fee --dry-run --generate-gnokey)

    _arguments -C \
        '1:verb or command:->first' \
        '*::args:->args'

    case "$state" in
        first)
            _describe 'verb' verbs
            _describe 'command' subcmds
            _values 'flags' $flags ;;
        args)
            case "${words[1]}" in
                completion)
                    _values 'shell' bash zsh fish ;;
            esac ;;
    esac
}

_gnopie`

const fishCompletion = `complete -c gnopie -n '__fish_use_subcommand' -a 'GET EVAL READ INSPECT CALL RUN' -d 'Verb'
complete -c gnopie -n '__fish_use_subcommand' -a 'config' -d 'Manage config'
complete -c gnopie -n '__fish_use_subcommand' -a 'completion' -d 'Shell completions'
complete -c gnopie -n '__fish_use_subcommand' -a 'version' -d 'Print version'
complete -c gnopie -n '__fish_seen_subcommand_from config' -a 'set get list'
complete -c gnopie -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
complete -c gnopie -l home -d 'Config home directory'
complete -c gnopie -l key -d 'Key name'
complete -c gnopie -l json -d 'JSON output'
complete -c gnopie -l quiet -d 'Suppress output'
complete -c gnopie -l send -d 'Coins to send'
complete -c gnopie -l gas-wanted -d 'Gas limit'
complete -c gnopie -l gas-fee -d 'Gas fee'
complete -c gnopie -l dry-run -d 'Simulate only'
complete -c gnopie -l generate-gnokey -d 'Print gnokey command'`
