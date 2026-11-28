> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `mdalert` - Markdown alerts

Render gnoweb-flavored Markdown alert blocks (note, tip, info, success, warning, caution) with optional title and folded mode.

## Usage

```go
import "gno.land/p/nt/mdalert/v0"

// One-liner helpers per type
md := mdalert.Warning("Heads up", "Disk almost full")

// Formatted variants accept ufmt-style args
md = mdalert.Infof("Stats", "%d users online", n)

// Full control via the Alert struct (e.g. folded by default)
a := mdalert.New(mdalert.TypeTip, "Click to expand", "Hidden details here", true)
md = a.String()
```

Rendered output (for the warning above):

```
> [!WARNING] Heads up
> Disk almost full
```

## API

```go
// Type identifies an alert variant.
type Type string

const (
    TypeCaution Type = "CAUTION"
    TypeInfo         = "INFO"
    TypeNote         = "NOTE"
    TypeSuccess      = "SUCCESS"
    TypeTip          = "TIP"
    TypeWarning      = "WARNING"
)

// Alert is a Markdown alert block.
type Alert struct {
    Type    Type   // Alert variant
    Title   string // Optional title (header line)
    Message string // Body; may contain newlines
    Folded  bool   // If true, render collapsed (only title visible)
}

// String renders the alert as Markdown. Returns "" if Type is empty or Message is blank (whitespace only).
func (a Alert) String() string

// New builds an Alert.
func New(t Type, title, msg string, folded bool) Alert

// Per-type helpers (unfolded). The *f variants format msg with ufmt.Sprintf.
func Caution(title, msg string) string
func Cautionf(title, format string, a ...any) string
func Info(title, msg string) string
func Infof(title, format string, a ...any) string
func Note(title, msg string) string
func Notef(title, format string, a ...any) string
func Success(title, msg string) string
func Successf(title, format string, a ...any) string
func Tip(title, msg string) string
func Tipf(title, format string, a ...any) string
func Warning(title, msg string) string
func Warningf(title, format string, a ...any) string
```

## Notes

- Alert types are documented in the Markdown docs realm: [/r/docs/markdown#alerts](/r/docs/markdown#alerts).
- Per-type helpers always render unfolded. For folded alerts use `New(...)` with `folded=true`.
- `title` and `msg` are emitted into Markdown as-is. When either carries untrusted input (a `Render(path)` segment, user text), wrap it with `sanitize.InlineText` from [`gno.land/p/nt/markdown/sanitize/v0`](../../markdown/sanitize/v0) first, or it can inject structure into the rendered page. The `*f` variants format via `ufmt.Sprintf`, which supports only ufmt's verb subset (no `%b`, `%o`, `%w`, `%+v`).
