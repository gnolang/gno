# `mdalert` - Markdown Alert Generator

A utility package for generating GitHub-style Markdown alerts with proper formatting. Supports different alert types with consistent styling for documentation and user interfaces.

## Features

- **Multiple alert types**: Note, Tip, Info, Warning, Caution, Success
- **Proper formatting**: GitHub-compatible Markdown alert syntax
- **Flexible content**: Support for multi-line content and custom messages
- **Type safety**: Predefined alert types prevent typos

## Supported Alert Types

- `NOTE` - General information
- `TIP` - Helpful suggestions  
- `INFO` - Informational content
- `WARNING` - Important warnings
- `CAUTION` - Critical warnings
- `SUCCESS` - Success messages

## Usage

```go
import "gno.land/p/nt/mdalert"

// Create simple alerts
noteAlert := mdalert.Note("This is important information")
tipAlert := mdalert.Tip("Here's a helpful tip")
warningAlert := mdalert.Warning("Be careful with this operation")

// Multi-line content
infoAlert := mdalert.Info("This is a detailed explanation.\n\nIt can span multiple lines.")

// Using alert types directly
alert := mdalert.New(mdalert.TypeSuccess, "Operation completed successfully!")
```

## Output Examples

The functions generate GitHub-style Markdown alerts:

```markdown
> [!NOTE]
> This is important information

> [!TIP]  
> Here's a helpful tip

> [!WARNING]
> Be careful with this operation

> [!SUCCESS]
> Operation completed successfully!
```

## API

### Alert Types
```go
const (
    TypeCaution Type = "CAUTION"
    TypeInfo         = "INFO"
    TypeNote         = "NOTE"
    TypeSuccess      = "SUCCESS"
    TypeTip          = "TIP"
    TypeWarning      = "WARNING"
)
```

### Functions
```go
// Generic alert creation
func New(alertType Type, content string) Alert

// Type-specific helpers
func Note(content string) Alert
func Tip(content string) Alert
func Info(content string) Alert
func Warning(content string) Alert
func Caution(content string) Alert
func Success(content string) Alert

// Convert to markdown
func (a Alert) Markdown() string
```

## Advanced Usage

```go
// Documentation generator
func GenerateDocumentation() string {
    var doc strings.Builder
    
    doc.WriteString("# API Documentation\n\n")
    
    // Add warnings about breaking changes
    warning := mdalert.Warning("This API is deprecated and will be removed in v2.0")
    doc.WriteString(warning.Markdown() + "\n\n")
    
    // Add helpful tips
    tip := mdalert.Tip("Use the new /v2/ endpoint for better performance")
    doc.WriteString(tip.Markdown() + "\n\n")
    
    // Add success notes
    success := mdalert.Success("All endpoints now support JSON responses")
    doc.WriteString(success.Markdown() + "\n\n")
    
    return doc.String()
}
```

## Integration with Contract Rendering

```go
func Render(path string) string {
    switch path {
    case "help":
        return renderHelp()
    case "warnings":
        return renderWarnings()
    default:
        return "Path not found"
    }
}

func renderHelp() string {
    var help strings.Builder
    
    help.WriteString("# Help Documentation\n\n")
    
    // Add informational alerts
    info := mdalert.Info("This contract manages user registrations")
    help.WriteString(info.Markdown() + "\n\n")
    
    tip := mdalert.Tip("Register early to get a lower user ID")
    help.WriteString(tip.Markdown() + "\n\n")
    
    return help.String()
}

func renderWarnings() string {
    var warnings strings.Builder
    
    warnings.WriteString("# System Warnings\n\n")
    
    if systemOverloaded() {
        caution := mdalert.Caution("System is experiencing high load")
        warnings.WriteString(caution.Markdown() + "\n\n")
    }
    
    if maintenanceScheduled() {
        warning := mdalert.Warning("Scheduled maintenance on Sunday 2PM UTC")
        warnings.WriteString(warning.Markdown() + "\n\n")
    }
    
    return warnings.String()
}
```

## Error and Status Reporting

```go
func ProcessUserAction(action string) string {
    switch action {
    case "register":
        if userExists() {
            warning := mdalert.Warning("User already registered")
            return warning.Markdown()
        }
        
        registerUser()
        success := mdalert.Success("Registration completed successfully!")
        return success.Markdown()
        
    case "delete":
        if !userExists() {
            info := mdalert.Info("No user found to delete")
            return info.Markdown()
        }
        
        caution := mdalert.Caution("This action cannot be undone")
        return caution.Markdown()
        
    default:
        note := mdalert.Note("Available actions: register, delete")
        return note.Markdown()
    }
}
```

## Best Practices

- **Appropriate types**: Use the right alert type for the content
  - `NOTE` for general information
  - `TIP` for helpful suggestions
  - `INFO` for detailed explanations
  - `WARNING` for important notices
  - `CAUTION` for critical warnings
  - `SUCCESS` for positive confirmations

- **Clear content**: Write concise, actionable messages
- **Consistent usage**: Use alerts consistently across your application
- **Multi-line support**: Break long content into readable paragraphs

## Dependencies

- `gno.land/p/nt/ufmt` - For string formatting

This package is perfect for creating professional documentation, user interfaces, and status reporting in Gno applications with properly formatted Markdown alerts.
