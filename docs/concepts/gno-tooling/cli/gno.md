---
id: gno-tooling-gno
---

# gno

`gno` is a handy tool for developing and prototyping Gno packages and realms. You may use `gno` to use the GnoVM without an actual blockchain to build or test realms in a local environment.

## Run `gno` Commands

The following command will run `gno`.

```bash
gno {SUB_COMMAND}
```

**Subcommands**

| Name         | Description                                |
| ------------ | ------------------------------------------ |
| `build`      | Builds a gno package.                      |
| `test`       | Tests a gno package.                       |
| `precompile` | Precompiles a `.gno` file to a `.go` file. |
| `repl`       | Starts a GnoVM REPL.                       |

### `build`

#### **Options**

| Name      | Type    | Description                                    |
| --------- | ------- | ---------------------------------------------- |
| `verbose` | Boolean | Displays extended information.                 |
| go-binary | String  | Go binary to use for building (default: `go`). |

### `test`

#### **Options**

| Name         | Type          | Description                                                        |
| ------------ | ------------- | ------------------------------------------------------------------ |
| `verbose`    | Boolean       | Displays extended information.                                     |
| `root-dir`   | String        | Clones location of github.com/gnolang/gno (gno tries to guess it). |
| `run`        | String        | Test name filtering pattern.                                       |
| `timeout`    | time.Duration | The maximum execution time in ns.                                  |
| `precompile` | Boolean       | Precompiles a `.gno` file to a `.go` file before testing.          |

### `precompile`

#### **Options**

| Name        | Type    | Description                                                     |
| ----------- | ------- | --------------------------------------------------------------- |
| `verbose`   | Boolean | Displays extended information.                                  |
| `skip-fmt`  | Boolean | Skips the syntax checking of generated `.go` files.             |
| `go-binary` | String  | The go binary to use for building (default: `go`).              |
| `go-binary` | String  | The gofmt binary to use for syntax checking (default: `gofmt`). |
| `output`    | String  | The output directory (default: `.`).                            |

### `repl`

#### **Options**

| Name       | Type    | Description                                                        |
| ---------- | ------- | ------------------------------------------------------------------ |
| `verbose`  | Boolean | Displays extended information.                                     |
| `root-dir` | String  | Clones location of github.com/gnolang/gno (gno tries to guess it). |
