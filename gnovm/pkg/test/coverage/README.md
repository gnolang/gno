# Gno Test Coverage Implementation

This document describes the code coverage implementation for Gno, which uses AST-based instrumentation to track test coverage.

## Overview

The coverage system implements code coverage through source code instrumentation, similar to other language's coverage implementation. It inserts tracking calls into the source code before execution to record which lines are executed during tests.

## Architecture

### Core Components

1. **CoverageTracker** (`coverage.go`)
   - Global singleton that maintains coverage data
   - Tracks execution counts for each line of code
   - Stores data as `map[filename]map[line]count`
   - Records both executable and executed lines

2. **CoverageInstrumenter** (`coverage.go`)
   - AST-based code transformer
   - Inserts `testing.MarkLine()` calls at key execution points
   - Handles all major Go/Gno language constructs
   - Automatically adds `testing` import when needed
   - Special handling for files containing the `cross` identifier

3. **Native Function Bindings** (`native.go`)
   - `X_markLine`: Native function that records line execution
   - `X_getCoverage`: Native function to retrieve coverage data
   - Exposed through the testing package as `testing.MarkLine()`

4. **Report Generation** (`report.go`)
   - Generates coverage reports in JSON and human-readable formats
   - Calculates coverage percentages per file and overall
   - Supports console output with detailed metrics

## Workflow

### 1. Test Execution with Coverage

When running `gno test -cover`:

```plain
User runs: gno test -cover ./package
    ↓
Test command initializes coverage tracker
    ↓
Package files are instrumented before compilation
    ↓
Tests execute instrumented code
    ↓
Coverage data is collected via MarkLine calls
    ↓
Coverage report is generated after tests complete
```

### 2. Code Instrumentation Process

The instrumentation process transforms source code by inserting tracking calls:

**Original Code:**

```go
func Add(a, b int) int {
    if a > 0 {
        return a + b
    }
    return b
}
```

**Instrumented Code:**

```go
import "testing"

func Add(a, b int) int {
    testing.MarkLine("path/to/file.gno", 1)
    if a > 0 {
        testing.MarkLine("path/to/file.gno", 2)
        testing.MarkLine("path/to/file.gno", 3)
        return a + b
    }
    testing.MarkLine("path/to/file.gno", 5)
    return b
}
```

### 3. Instrumentation Points

The system instruments the following AST nodes:

- Function declarations (start of function body)
- If statements (condition, then and else branches)
- For/Range loops (condition and loop body)
- Switch/Select statements (case bodies)
- Return statements (before return)
- Case clauses (at the start of each case)
- Block statements containing returns

## Current Implementation Status

### ✅ Working Features

1. **Pure Package Coverage** (`p/`)
   - Full coverage support for pure packages
   - Fixed issue where packages were showing 0% coverage due to re-loading
   - Correctly preserves instrumentation throughout test execution

2. **Basic Realm Coverage** (`r/`)
   - Coverage works for simple realm packages
   - Realm dependencies are loaded when needed for coverage

3. **Coverage Reporting**
   - JSON output with detailed line-by-line data
   - Human-readable console output with percentages
   - File-level and package-level metrics

### ⚠️ Known Limitations

1. **Cross-Realm Calls**
   - Files containing the `cross` identifier are skipped
   - This prevents preprocessing panics but results in 0% coverage
   - Affects realms using cross-realm function calls

2. **Realm Dependencies**
   - Complex realm interdependencies may still cause issues
   - Some edge cases with circular dependencies

3. **Performance Impact**
   - Instrumentation adds overhead to test execution
   - More noticeable with large test suites

4. **Coverage Scope**
   - Only line coverage is implemented (no branch coverage)
   - Imported packages are not instrumented (by design)

## Usage

### Basic Commands

```bash
# Run tests with coverage
gno test -cover ./package

# Verbose mode (shows instrumented code)
gno test -cover -v ./package

# Generate coverage profile
gno test ./package -cover -coverprofile=coverage.json
```

### Coverage Report Format

#### JSON Output

```json
{
  "files": {
    "package/file.gno": {
      "lines": {
        "10": 5,    // line 10 executed 5 times
        "11": 0,    // line 11 not executed
        "12": 3     // line 12 executed 3 times
      },
      "total": 50,
      "covered": 35,
      "coverage": 70.0
    }
  }
}
```

#### Console Output

```plain
Coverage Report:
Total Lines: 150
Covered Lines: 120
Overall Coverage: 80.00%

File: package/file.gno
  Total Lines: 50
  Covered Lines: 40
  Coverage: 80.00%
```

## Implementation Details

### Package Loading and Import Resolution

The coverage system modifies the normal package loading process:

1. **Instrumentation First**: Package files are instrumented before any other processing
2. **Skip Re-loading**: The tested package is not re-loaded during import resolution
3. **Preserve Instrumentation**: Ensures instrumented code is used throughout execution
4. **Realm Dependencies**: Optionally loads realm dependencies when testing realms

### Special Handling for Cross

Files containing the `cross` identifier require special treatment:

- The entire file is skipped during instrumentation
- Executable lines are still registered for accurate metrics
- Prevents preprocessing errors with this special identifier
- Results in 0% coverage for affected files

### Native Function Integration

Coverage functions are implemented as native Go functions:

- `X_markLine(filename, line)`: Records line execution
- `X_getCoverage()`: Returns coverage data as JSON
- Integrated into Gno VM's native function system
- Minimal performance overhead

## Future Improvements

### High Priority

- [ ] Fix cross-realm call instrumentation issues
- [X] Add support for coverage output file (`-coverprofile`)
- [ ] Implement HTML report generation
- [ ] Add branch coverage support

### Medium Priority

- [ ] Coverage merging from multiple test runs
- [ ] Coverage thresholds with test failure option
- [ ] Exclude patterns for files/directories
- [ ] Integration with standard Go coverage tools

### Low Priority

- [ ] Function-level coverage metrics
- [ ] Coverage diff between commits
- [ ] Performance optimizations for large codebases
- [ ] Coverage trend analysis

## Troubleshooting

### Common Issues

**Cross-Realm Call Errors**

- Cause: `cross` identifier cannot be instrumented
- Workaround: Files with `cross` are skipped
