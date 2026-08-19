# stdlibs_diff

stdlibs_diff is a tool that generates an html report indicating differences between gno standard libraries and go standard libraries.

## Usage

Compare the `go` standard libraries the `gno` standard libraries

```shell
./stdlibs_diff -src <path to go standard libraries> -dst <path to gno standard libraries> -out <output directory>
```

Compare the `gno` standard libraries the `go` standard libraries

```shell
./stdlibs_diff -src <path to gno standard libraries> -dst <path to go standard libraries> -out <output directory>
```


## Parameters

| Flag       | Description                                                        | Default value |
| ---------- | ------------------------------------------------------------------ | ------------- |
| src        | Directory containing packages that will be compared to destination | None          |
| dst        | Directory containing packages; used to compare src packages        | None          |
| out        | Directory where the report will be created                         | None          |

## Baseline Go version

In CI (`.github/workflows/deploy-pages.yml`) the report is generated against a
**pinned Go 1.26.1**, not against the toolchain in the top-level `go.mod`. The
report is a drift measurement, so the version it compares against is an
editorial choice: letting it follow `go.mod` would silently re-baseline the
report every time the build toolchain is bumped, quietly erasing drift that had
previously been visible.

The workflow asserts the version it resolved and fails loudly on a mismatch. To
bump the baseline, change `go-version` **and** the assertion in that workflow,
and update this section.

Locally, `make gen` still uses whatever `go env GOROOT` reports. To reproduce
CI exactly, point it at a matching toolchain:

```shell
GOROOT_SAVE=$(go1.26.1 env GOROOT) make gen
```

### Known intentional divergences

Some packages are deliberately *not* sourced from the baseline and will always
appear as large diffs. These are expected, not regressions:

| Path | Sourced from | Why |
| ---- | ------------ | --- |
| `unicode/{tables,letter,graphic,casetables,digit}.gno` | go1.27rc2 | Unicode 17.0.0; the baseline still ships Unicode 15.0.0 |
| `strconv/isprint.gno` | go1.27rc2 | generated from the same tables, must match `unicode` |
| `math/rand/**` | `math/rand/v2` | gno ports v2, not v1 |
| `strconv/{atof,atoi,ftoa,decimal,...}.gno` | `internal/strconv` | Go 1.26 moved these out of the public package |

## Tips

An index.html is generated at the root of the report location. Utilize it to navigate easily through the report.
