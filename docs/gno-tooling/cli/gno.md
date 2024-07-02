---
id: gno-tooling-gno
---

# gno

`gno` is a handy tool for developing and prototyping Gno packages and realms. You may use `gno` to use the GnoVM without an actual blockchain to build or test realms in a local environment.

## Run `gno` Commands

The following command will run `gno`, modified by the sub command.

```bash
gno {SUB_COMMAND}
```
To print a comprehensive list of `gno` commands and a brief description of each command, use the command `gno`.

**Sub Commands**

| Sub Command        | Description                                | Usage                                            |
| ------------------ | ------------------------------------------ | ------------------------------------------------ |
| `mod`              | Manage gno.mod.                            | `gno mod <sub_command>`                          |
| `test`             | Runs the tests for the specified packages. | `gno test [flags] <package> [<package>...]`      |
| `lint`             | Runs the linter for the specified packages.| `gno lint [flags] <package> [<package>...]`      |
| `run`              | Runs the specified `.gno` file(s).         | `gno run [flags] <file> [<file>...]`             |
| `transpile`        | Transpiles a `.gno` file to a `.go` file.  | `gno transpile [flags] <package> [<package>...]` |
| `repl`             | Starts a GnoVM REPL.                       | `gno repl`                                       |
| `doc`              | Get documentation for the specified package or symbol (type, function, method, or variable/constant). | `gno doc <package/symbol>` |
| `env`              | Prints Gno environment information.        | `gno env`                                        |
| `bug`              | Starts a bug report in GitHub.             | `gno bug`                                        |

### `mod`

#### **Sub Commands**

| Name         | Description                                                        |
| ------------ | ------------------------------------------------------------------ |
| `download`   | Downloads modules to local cache.                                  |
| `init`       | Initialize `gno.mod` file in current directory.                    |
| `tidy`       | Add missing modules and remove unused modules.                     |
| `why`        | Explains why modules are needed.                                   |

### `test`

#### **Flags**

| Name         | Type          | Description                                                        |
| ------------ | ------------- | ------------------------------------------------------------------ |
| `v`          | Boolean       | Displays verbose output.                                           |
| `root-dir`   | String        | Clones location of github.com/gnolang/gno (gno tries to guess it). |
| `run`        | String        | Test name filtering pattern.                                       |
| `timeout`    | time.Duration | The maximum execution time in ns.                                  |
| `transpile`  | Boolean       | Transpiles a `.gno` file to a `.go` file before testing.           |

### `transpile`

#### **Flags**

| Name              | Type    | Description                                                       |
| ----------------- | ------- | ----------------------------------------------------------------- |
| `go-binary`       | String  | The go binary to use for building (default: `go`).                |
| `go-fmt-binary`   | String  | The `gofmt` binary to use for syntax checking (default: `gofmt`). |
| `gobuild`         | Boolean | Run `go build` on generated `.go` files, ignoring test files.     |     
| `output`          | String  | The output directory (default: `.`).                              |
| `skip-fmt`        | Boolean | Skips the syntax checking of generated `.go` files.               |
| `skip-imports`    | Boolean | Do not transpile imports recursively.                             |
| `v`               | Boolean | Displays verbose output.                                          |

### `repl`

#### **Options**

| Name       | Type    | Description                                                        |
| ---------- | ------- | ------------------------------------------------------------------ |
| `root-dir` | String  | Clones location of github.com/gnolang/gno (gno tries to guess it). |
