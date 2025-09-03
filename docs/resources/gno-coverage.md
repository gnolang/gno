# Code Coverage in Gno

Gno supports code coverage analysis similar to Go's coverage tools. Coverage analysis helps you understand which parts of your code are executed during tests.

## Basic Usage

Enable coverage analysis with the `-cover` flag:

```bash
gno test -cover ./...
```

This will run your tests and report the percentage of statements that were executed.

## Coverage Modes

Gno supports three coverage modes, controlled by the `-covermode` flag:

### Set Mode (Default)
```bash
gno test -covermode set ./...
```
- Tracks whether each statement was executed at least once
- Boolean coverage: each statement is either covered or not covered
- Fastest mode with minimal overhead

### Count Mode
```bash
gno test -covermode count ./...
```
- Tracks how many times each statement was executed
- Useful for identifying hot paths and understanding execution patterns
- Slightly more overhead than set mode

### Atomic Mode
```bash
gno test -covermode atomic ./...
```
- Like count mode, but uses atomic operations for thread safety
- Use when running tests in parallel or with concurrent code
- Highest overhead but thread-safe

## Coverage Profiles

Generate a coverage profile file with the `-coverprofile` flag:

```bash
gno test -coverprofile coverage.out ./...
```

The coverage profile can be used with Go's coverage tools:

```bash
# View coverage in the browser (if Go tools are available)
go tool cover -html=coverage.out

# Generate text report
go tool cover -func=coverage.out
```

## Examples

### Basic coverage report
```bash
$ gno test -cover ./examples/gno.land/p/demo/ufmt
coverage: 85.7% of statements in gno.land/p/demo/ufmt
ok      ./examples/gno.land/p/demo/ufmt    0.12s
```

### Count mode with profile
```bash
$ gno test -covermode count -coverprofile coverage.out ./examples/gno.land/p/demo/ufmt
coverage: 85.7% of statements in gno.land/p/demo/ufmt
ok      ./examples/gno.land/p/demo/ufmt    0.15s
```

### Multiple packages
```bash
$ gno test -cover ./examples/gno.land/p/demo/...
coverage: 92.3% of statements in gno.land/p/demo/avl
coverage: 85.7% of statements in gno.land/p/demo/ufmt
coverage: 78.9% of statements in gno.land/p/demo/json
...
```

## Integration with CI

Coverage can be integrated into CI workflows. For example, in GitHub Actions:

```yaml
- name: Run tests with coverage
  run: gno test -covermode count -coverprofile coverage.out ./...

- name: Upload coverage reports
  uses: codecov/codecov-action@v3
  with:
    file: coverage.out
```

## Limitations

- Coverage tracking adds some runtime overhead
- Coverage is tracked at the statement level, not line level
- Some complex control flow may not be tracked perfectly
- Coverage data is collected per package, not globally

## Implementation Notes

Gno's coverage implementation tracks statement execution in the GnoVM during test runs. Unlike Go's coverage which rewrites source code, Gno adds coverage tracking directly to the virtual machine execution engine.

This approach provides accurate coverage data while maintaining the integrity of the original source code during testing.
