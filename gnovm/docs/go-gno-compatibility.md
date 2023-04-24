# Go<>Gno compatibility

WIP: does not reflect the current state yet.

## Native keywords

| keyword             | status   |
|---------------------|----------|
| switch              | OK       |
| go                  | missing  |
| chan                | missing  |

## Stdlibs

| package             | status   |
|---------------------|----------|
| foo                 | partial  |
| bar                 | missing  |
| baz                 | full     |
| foobarbaz           | outdated |

## Tooling (`gno` binary)

| go command      | gno command      | comment |
|-----------------|------------------|---------|
| go build        | gno build        | partial |
| go mod download | gno mod download | OK      |
| go help         | n/a              | missing |

