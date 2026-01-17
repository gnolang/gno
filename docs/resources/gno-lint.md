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
| `-v` | false | Verbose output |
| `--mode` | default | Lint mode: `default`, `strict`, `warn-only` |
| `--format` | text | Output format: `text`, `json` |
| `--list-rules` | false | List available rules and exit |
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

### AVL001: Unbounded AVL Tree Iteration

**Severity:** Warning
**Since:** v1.0.0
**Category:** AVL

Detects calls to `avl.Tree.Iterate()` or `avl.Tree.ReverseIterate()` with empty
string bounds (`"", ""`), which iterates over the entire tree and may cause
performance issues or denial of service.

**Example (problematic):**
```go
tree.Iterate("", "", func(key string, value interface{}) bool {
    // Iterates ALL entries - potentially expensive
    return false
})
```

**Example (correct):**
```go
// Option 1: Use bounds
tree.Iterate("a", "z", func(key string, value interface{}) bool {
    return false
})

// Option 2: Use IterateByOffset with limit
tree.IterateByOffset(0, 100, func(key string, value interface{}) bool {
    return false
})
```

**Rationale:** Unbounded iteration on large trees can be expensive and may allow
malicious actors to trigger denial of service by populating trees with many entries.

---

### GLOBAL001: Exported Package-Level Variable

**Severity:** Warning
**Since:** v1.0.0
**Category:** General

Detects exported (uppercase) package-level `var` declarations, which may indicate
poor encapsulation.

**Example (problematic):**
```go
var Counter int  // Exported global - can be modified from anywhere
```

**Example (correct):**
```go
var counter int  // Unexported - controlled access

func GetCounter() int {
    return counter
}

func IncrementCounter(_ realm) {
    counter++
}
```

**Rationale:** Exported globals can be modified by any code that imports the package,
making it harder to reason about state changes and maintain invariants.

---

## Suppressing Issues

Use `//nolint` comments to suppress specific issues:

```go
//nolint:AVL001
tree.Iterate("", "", func(key string, value interface{}) bool {
    // Intentionally unbounded - we know the tree is small
    return false
})

//nolint:GLOBAL001
var Config = DefaultConfig()  // Intentionally exported

//nolint  // Suppress all rules for next line
var GlobalState int
```

### Nolint Placement

- Place the comment on the line **above** the issue
- Multiple rules: `//nolint:AVL001,GLOBAL001`
- All rules: `//nolint`

---

## Output Formats

### Text (default)

```
counter.gno:42:5: warning: unbounded Iterate on avl.Tree (AVL001)
counter.gno:10:5: warning: exported package-level variable: Counter (GLOBAL001)

Found 2 issue(s): 0 error(s), 2 warning(s), 0 info
```

### JSON

```bash
gno lint --format=json ./mypackage
```

```json
[
  {
    "ruleId": "AVL001",
    "severity": "warning",
    "message": "unbounded Iterate on avl.Tree",
    "filename": "counter.gno",
    "line": 42,
    "column": 5
  }
]
```

---

## Configuration (Future)

> **Note:** Configuration file support is planned for a future release.

A `gnolint.toml` file will allow per-project configuration:

```toml
[lint]
mode = "default"
format = "text"
disable = ["GLOBAL001"]

[lint.nolint]
require_reason = false
```

---

## Adding New Rules (Contributor Guide)

### Rule Structure

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
        ID:       "CAT001",       // Category + number
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

### Rule Categories

| Category | Prefix | Description |
|----------|--------|-------------|
| `CategoryAVL` | AVL | AVL tree usage issues |
| `CategoryGeneral` | GLOBAL | General code issues |

### RuleContext Fields

| Field | Type | Description |
|-------|------|-------------|
| `File` | `*gnolang.FileNode` | Current file being analyzed |
| `Source` | `string` | Raw source code |
| `Parents` | `[]gnolang.Node` | Parent node stack (innermost last) |

### Best Practices

1. **Keep rules focused** - One concern per rule
2. **Use descriptive IDs** - `AVL001` not `RULE1`
3. **Provide helpful messages** - Include what's wrong and how to fix
4. **Handle edge cases** - Check for nil, empty, etc.
5. **Add tests** - Both unit tests and txtar integration tests

### Testing Your Rule

1. **Unit test** in `rules/myrule_test.go`:
```go
func TestMyRule(t *testing.T) {
    // Table-driven tests
}
```

2. **Integration test** in `gnovm/cmd/gno/testdata/lint/`:
```
# Test MyRule detection

! gno lint .

stderr 'CAT001'
stderr 'expected error message'

-- main.gno --
package main

// Code that triggers the rule

-- gnomod.toml --
module = "gno.land/r/test/myrule"
gno = "0.9"
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (no errors, warnings allowed in default mode) |
| 1 | Issues found (errors, or any issues in strict mode) |

---

## See Also

- [Configuring Gno Projects](./configuring-gno-projects.md)
- [Gno Testing](./gno-testing.md)
- [Effective Gno](./effective-gno.md)
