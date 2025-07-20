# `gnokeykc`: CLI tool enhancing gnokey for system keychain integration

`gnokeykc` is a Go-based CLI tool that enhances [`gnokey`](../../gno.land/cmd/gnokey) by integrating with your system's keychain. It adds `gnokey kc ...` subcommands to set and unset passwords in the keychain, allowing Gnokey to fetch passwords directly from the keychain instead of prompting for terminal input.

## Usage

    gnokey kc -h

## Terminal Alias

For ease of use, set up a terminal alias to replace `gnokey` with `gnokeykc`:

    echo "alias gnokey='gnokeykc'" >> ~/.bashrc && source ~/.bashrc

Now, `gnokey` commands will use `gnokeykc`, fetching passwords from the keychain.

