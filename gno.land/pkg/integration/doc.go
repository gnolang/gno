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
// 1. `gnoland [start|stop]`:
//   - The gnoland node doesn't start automatically. This enables the user to do some
//     pre-configuration or pass custom arguments to the start command.
//
// 2. `gnokey`:
//   - Supports most of the common commands.
//   - `--remote`, `--insecure-password-stdin`, and `--home` flags are set automatically to
//     communicate with the gnoland node.
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
//   - LOG_LEVEL:
//     The logging level to be used, which can be one of "error", "debug", "info", or an empty string.
//     If empty, the log level defaults to "debug".
//
//   - LOG_DIR:
//     If set, logs will be directed to the specified directory.
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
//   - USER_SEED_test1:
//     Contains the seed for the test1 account.
//
//   - USER_ADDR_test1:
//     Contains the address for the test1 account.
//
//   - RPC_ADDR:
//     Points to the gnoland node's remote address. It's set only if the node has started.
//
// For a more comprehensive guide on original behaviors, additional commands and environment
// variables, refer to the original documentation of testscripts available here:
// https://github.com/rogpeppe/go-internal/blob/master/testscript/doc.go
package integration
