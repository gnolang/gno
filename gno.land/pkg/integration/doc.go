// Package integration offers utilities to run txtar-based tests against the gnoland system
// by extending the functionalities provided by the standard testscript package. This package is
// currently in an experimental phase and may undergo significant changes in the future.
//
// SetupGnolandTestScript, sets up the environment for running txtar tests, introducing additional
// commands like "gnoland" and "gnokey" into the test script ecosystem. Specifically, it allows the
// user to initiate an in-memory gnoland node and interact with it via the `gnokey` command.
//
// Additional Command Overview:
//
// 1. `gnoland [start|stop|restart]`:
//   - The gnoland node doesn't start automatically. This enables the user to do some
//     pre-configuration or pass custom arguments to the start command.
//   - `gnoland restart` will simulate restarting a node, as in stopping and
//     starting it again, recovering state from the persisted database data.
//   - `gnoland start -non-validator` can be used to start a node as a non-validator node.
//
// 2. `gnokey`:
//   - Supports most of the common commands.
//   - `--remote`, `--insecure-password-stdin`, and `--home` flags are set automatically to
//     communicate with the gnoland node.
//   - In order to handle escape sequences like `\n` within arguments, you can enclose the argument
//     in `"`
//
// 3. `adduser`:
//   - Must be run before `gnoland start`.
//   - Creates a new user in the default keybase directory. ( Optionally, an std.Coins string can be provided, e.g. `1337ugnot,42foo,24bar` )
//
// 4. `adduserfrom`:
//   - Must be run before `gnoland start`.
//   - Creates a new user in the default keybase directory from a given seed. ( Optionally, account and index can be provided )
//
// 5. `loadpkg`:
//   - Must be run before `gnoland start`.
//   - Loads a specific package from the 'examples' directory or from the working ($WORK) directory.
//   - Can be used to load a single package or all packages within a directory.
//   - If the target package has a `gnomod.toml`, all its dependencies (and their respective
//     dependencies) will also be loaded.
//   - The command takes either one or two arguments. The first argument is the name of the package(s),
//     and the second (optional) argument is the path to the package(s).
//     Examples:
//     -- # Load a package from the 'examples' directory:
//     -- loadpkg gno.land/p/nt/ufmt/v0
//     -- # Load a package `./bar` from the testscript's working directory with the name `gno.land/r/foobar/bar`:
//     -- loadpkg gno.land/r/foobar/bar $WORK/bar
//   - If the path is not prefixed with the working directory, it is assumed to be relative to the
//     examples directory.
//   - It's important to note that the load order is significant when using multiple `loadpkg`
//     command; packages should be loaded in the order they are dependent upon.
//
// 6. `patchpkg`:
//   - Patches any loaded files by package by replacing all occurrences of the first argument with the second.
//   - This is mostly used to replace hardcoded addresses from loaded packages.
//   - NOTE: this command may only be temporary, as it's not best approach to
//     solve the above problem
//
// Logging:
//
// Gnoland logs aren't forwarded to stdout to avoid overwhelming the tests with too much
// information. Instead, a log directory can be specified with `LOG_DIR`, or you
// can set `TESTWORK=true`
// to persist logs in the txtar working directory. In any case, the log file should be printed
// on start if one of these environment variables is set.
//
// Accounts:
//
// By default, only the test1 user will be created in the default keybase directory,
// with no password set. The default gnoland genesis balance file and the genesis
// transaction file are also registered by default.
//
// Examples:
//
// Examples can be found in the `testdata` directory of this package.
//
// Environment Variables:
//
// Input:
//
//   - TESTWORK:
//     A boolean that, when enabled, retains working directories after tests for
//     inspection. If enabled, gnoland logs will be persisted inside this
//     folder.
//
//   - UPDATE_SCRIPTS:
//     A boolean that, when enabled, updates the test scripts if a `cmp` command
//     fails and its second argument refers to a file inside the testscript
//     file. The content will be quoted with txtar.Quote if needed, requiring
//     manual edits if it's not unquoted in the script.
//
// Output (available inside testscripts files):
//
//   - WORK:
//     The path to the temporary work directory tree created for each script.
//
//   - GNOROOT:
//     Points to the local location of the gno repository, serving as the GOROOT equivalent for gno.
//
//   - GNOHOME:
//     Refers to the local directory where gnokey stores its keys.
//
//   - GNODATA:
//     The path where the gnoland node stores its configuration and data. It's
//     set only if the node has started.
//
//   - xxx_user_seed:
//     Where `xxx` is the account name; Contains the seed for the test1 account.
//
//   - xxx_user_addr:
//     Where `xxx` is the account name; Contains the address for the test1 account.
//
//   - xxx_account_num:
//     Where `xxx` is the account name; Contains the account number for the test1 account.
//
//   - xxx_account_seq:
//     Where `xxx` is the account name; Contains the address for the test1 account.
//
//   - RPC_ADDR:
//     Points to the gnoland node's remote address. It's set only if the node has started.
//
// For a more comprehensive guide on original behaviors, additional commands and environment
// variables, refer to the original documentation of testscripts available here:
// https://github.com/rogpeppe/go-internal/blob/master/testscript/doc.go
package integration
