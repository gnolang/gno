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
| `test`       | Tests a gno package.                       |
| `transpile`  | Transpiles a `.gno` file to a `.go` file. |
| `repl`       | Starts a GnoVM REPL.                       |

### `test`

#### **Options**

| Name         | Type          | Description                                                        |
| ------------ | ------------- | ------------------------------------------------------------------ |
| `v`          | Boolean       | Displays verbose output.                                     |
| `root-dir`   | String        | Clones location of github.com/gnolang/gno (gno tries to guess it). |
| `run`        | String        | Test name filtering pattern.                                       |
| `timeout`    | time.Duration | The maximum execution time in ns.                                  |
| `transpile`  | Boolean       | Transpiles a `.gno` file to a `.go` file before testing.          |

### `transpile`

#### **Options**

| Name        | Type    | Description                                                     |
| ----------- | ------- | --------------------------------------------------------------- |
| `v`         | Boolean | Displays verbose output.                                  |
| `skip-fmt`  | Boolean | Skips the syntax checking of generated `.go` files.             |
| `gobuild`   | Boolean | Run `go build` on generated `.go` files, ignoring test files.   |
| `go-binary` | String  | The go binary to use for building (default: `go`).              |
| `gofmt`     | String  | The gofmt binary to use for syntax checking (default: `gofmt`). |
| `output`    | String  | The output directory (default: `.`).                            |

### `repl`

#### **Options**

| Name       | Type    | Description                                                        |
| ---------- | ------- | ------------------------------------------------------------------ |
| `root-dir` | String  | Clones location of github.com/gnolang/gno (gno tries to guess it). |
