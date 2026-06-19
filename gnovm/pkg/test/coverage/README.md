# VM-Level Coverage Implementation

This package implements VM-level code coverage tracking for Gno tests, providing a decoupled approach that tracks execution at the virtual machine level rather than through AST instrumentation.

## Features

- **VM-level tracking**: Tracks code execution during VM runtime without modifying source files
- **Decoupled design**: Uses interfaces to minimize coupling with the Machine struct
- **File-by-file coverage**: Detailed coverage information for each source file
- **ANSI visualization**: Terminal-based coverage visualization with color coding

## Usage

### Running tests with coverage

```bash
gno test -cover ./path/to/package
```

Or, 

```bash
go run ./gnovm/cmd/gno test -cover <pkg_path> -show <file_name>
```

### Viewing coverage visualization

```bash
gno test -cover -show "*.gno" ./path/to/package
```

### Coverage output formats

- Console output: Default coverage summary printed to stderr
- File output: Use `-coverprofile=coverage.out` to save coverage data

## Architecture

The implementation consists of several key components:

1. **CoverageTracker Interface** (`coverage_interface.go`): Defines the contract for coverage tracking
2. **VM Tracker** (`tracker.go`): Implements the coverage tracking logic
3. **Analyzer** (`analyzer.go`): Analyzes packages to identify executable lines
4. **Report** (`report.go`): Generates coverage reports
5. **Visualizer** (`visualize.go`): Provides ANSI-colored terminal visualization

## Color Coding

When using the `-show` flag, the coverage visualization uses:
- 🟢 Green: Executed lines
- 🔴 Red: Not executed lines
- 🔘 Gray: Non-executable lines (comments, blank lines)
- ⚪ White: Line numbers

## Implementation Details

The coverage tracking works by:
1. Registering executable lines during package analysis
2. Tracking execution in `doOpExec` and `doOpEval` VM operations
3. Collecting coverage data without modifying the source code
4. Generating reports and visualizations based on collected data
