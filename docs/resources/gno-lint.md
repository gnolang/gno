# Gno Lint

`gno lint` is a static analysis tool for Gno code that helps identify potential
issues, enforce best practices, and maintain code quality.

## Quick Start

```bash
# Lint a package
gno lint ./mypackage

# Lint with strict mode (warnings become errors)
gno lint --mode=strict ./mypackage

# List available rules
gno lint --list-rules
```

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-C` | - | Change to directory before running command |
| `-v` | false | Verbose output |
| `--mode` | default | Lint mode: `default`, `strict`, `warn-only` |
| `--format` | text | Output format: `text`, `json` |
| `--list-rules` | false | List available rules and exit |
| `--disable-rules` | - | Comma-separated list of rules to disable (e.g., `AVL001,GLOBAL001`) |
| `--root-dir` | auto | Gno root directory |
| `--auto-gnomod` | true | Auto-generate gnomod.toml if missing |

### Lint Modes

| Mode | Behavior | Exit Code |
|------|----------|-----------|
| `default` | Warnings + errors reported | 1 if errors, 0 if only warnings |
| `strict` | All issues become errors | 1 if any issues |
| `warn-only` | All issues become warnings | Always 0 |

### Severity Levels

| Level | Description | Exit Code (default mode) |
|-------|-------------|--------------------------|
| `info` | Informational messages | 0 |
| `warning` | Potential issues that may need attention | 0 |
| `error` | Issues that must be fixed | 1 |

## Available Rules

- **AVL001** (warning): Unbounded AVL tree iteration - detects `avl.Tree.Iterate()` or `ReverseIterate()` with empty bounds (`"", ""`)
- **GLOBAL001** (warning): Exported package-level variable - detects exported (uppercase) `var` declarations at package level

## Suppressing Issues

Use `//nolint` comments to suppress specific issues:

```go
//nolint:AVL001
tree.Iterate("", "", func(key string, value interface{}) bool {
    return false
})

//nolint:GLOBAL001
var Config = DefaultConfig()

//nolint  // Suppress all rules for next line
var GlobalState int
```

### Nolint Placement

- Place the comment on the line **above** the issue
- Multiple rules: `//nolint:AVL001,GLOBAL001`
- All rules: `//nolint`

## Output Formats

### Text (default)

```
counter.gno:10:5: warning: exported package-level variable: Counter (GLOBAL001)

Found 1 issue(s): 0 error(s), 1 warning(s), 0 info
```

### JSON

```bash
gno lint --format=json ./mypackage
```

```json
[
  {
    "ruleId": "GLOBAL001",
    "severity": "warning",
    "message": "exported package-level variable: Counter",
    "filename": "counter.gno",
    "line": 10,
    "column": 5
  }
]
```

## Adding New Rules

Rules live in `gnovm/pkg/lint/rules/` and implement the `lint.Rule` interface:

```go
package rules

import (
    "github.com/gnolang/gno/gnovm/pkg/gnolang"
    "github.com/gnolang/gno/gnovm/pkg/lint"
)

func init() {
    lint.MustRegister(&MyRule{})
}

type MyRule struct{}

func (MyRule) Info() lint.RuleInfo {
    return lint.RuleInfo{
        ID:       "CAT001",
        Category: lint.CategoryXxx,
        Name:     "my-rule-name",
        Severity: lint.SeverityWarning,
    }
}

func (MyRule) Check(ctx *lint.RuleContext, node gnolang.Node) []lint.Issue {
    // Check node and return issues
    return nil
}
```

### RuleContext Fields

| Field | Type | Description |
|-------|------|-------------|
| `File` | `*gnolang.FileNode` | Current file being analyzed |
| `Source` | `string` | Raw source code |
| `Parents` | `[]gnolang.Node` | Parent node stack (innermost last) |

### Testing Your Rule

1. Unit test in `rules/myrule_test.go`
2. Integration test in `gnovm/cmd/gno/testdata/lint/`

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (no errors, warnings allowed in default mode) |
| 1 | Issues found (errors, or any issues in strict mode) |
