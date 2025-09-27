# `combinederr` - Combined Error Handling

A utility package for combining multiple errors into a single error object. Useful for collecting and reporting multiple validation errors or operation failures.

## Features

- **Multiple errors**: Combine several errors into one
- **Clean formatting**: Automatic semicolon-separated error messages
- **Nil handling**: Safely handles nil errors (skips them)
- **Simple API**: Easy to use with familiar error patterns

## Usage

```go
import "gno.land/p/nt/combinederr"

// Create combined error
var combined combinederr.CombinedError

// Add individual errors
err1 := errors.New("first error")
err2 := errors.New("second error")
err3 := errors.New("third error")

combined.Add(err1)
combined.Add(err2)
combined.Add(err3)

// Get combined error message
fmt.Println(combined.Error())
// Output: "first error; second error; third error"

// Check if there are any errors
if combined.HasErrors() {
    return &combined
}
```

## Validation Example

```go
func ValidateUser(user *User) error {
    var errors combinederr.CombinedError
    
    if user.Name == "" {
        errors.Add(fmt.Errorf("name is required"))
    }
    
    if user.Email == "" {
        errors.Add(fmt.Errorf("email is required"))
    } else if !isValidEmail(user.Email) {
        errors.Add(fmt.Errorf("email format is invalid"))
    }
    
    if user.Age < 0 {
        errors.Add(fmt.Errorf("age cannot be negative"))
    }
    
    if user.Age > 150 {
        errors.Add(fmt.Errorf("age seems unrealistic"))
    }
    
    if errors.HasErrors() {
        return &errors
    }
    
    return nil
}

// Usage
user := &User{Name: "", Email: "invalid", Age: -5}
if err := ValidateUser(user); err != nil {
    fmt.Println(err.Error())
    // Output: "name is required; email format is invalid; age cannot be negative"
}
```

## API

```go
type CombinedError struct {
    // private fields
}

// Add an error to the collection
func (e *CombinedError) Add(err error)

// Get combined error message
func (e *CombinedError) Error() string

// Check if any errors were added
func (e *CombinedError) HasErrors() bool

// Get count of errors
func (e *CombinedError) Count() int
```

## Batch Operations

```go
func ProcessBatch(items []Item) error {
    var errors combinederr.CombinedError
    
    for i, item := range items {
        if err := processItem(item); err != nil {
            // Add context to error
            contextErr := fmt.Errorf("item %d: %w", i, err)
            errors.Add(contextErr)
        }
    }
    
    if errors.HasErrors() {
        return fmt.Errorf("batch processing failed: %w", &errors)
    }
    
    return nil
}
```

## Configuration Validation

```go
func ValidateConfig(config *Config) error {
    var errors combinederr.CombinedError
    
    // Validate database settings
    if config.Database.Host == "" {
        errors.Add(fmt.Errorf("database host is required"))
    }
    
    if config.Database.Port <= 0 {
        errors.Add(fmt.Errorf("database port must be positive"))
    }
    
    // Validate API settings
    if config.API.Port <= 0 || config.API.Port > 65535 {
        errors.Add(fmt.Errorf("API port must be between 1 and 65535"))
    }
    
    if config.API.Secret == "" {
        errors.Add(fmt.Errorf("API secret is required"))
    }
    
    // Validate feature flags
    if config.Features.EnableAuth && config.Auth.Provider == "" {
        errors.Add(fmt.Errorf("auth provider is required when auth is enabled"))
    }
    
    if errors.HasErrors() {
        return &errors
    }
    
    return nil
}
```

## Error Accumulation Pattern

```go
type DataProcessor struct {
    errors combinederr.CombinedError
}

func (dp *DataProcessor) ProcessRecord(record Record) {
    if err := dp.validateRecord(record); err != nil {
        dp.errors.Add(err)
        return
    }
    
    if err := dp.transformRecord(record); err != nil {
        dp.errors.Add(err)
        return
    }
    
    if err := dp.storeRecord(record); err != nil {
        dp.errors.Add(err)
        return
    }
}

func (dp *DataProcessor) GetErrors() error {
    if dp.errors.HasErrors() {
        return &dp.errors
    }
    return nil
}

// Usage
processor := &DataProcessor{}
for _, record := range records {
    processor.ProcessRecord(record)
}

if err := processor.GetErrors(); err != nil {
    fmt.Printf("Processing failed: %v\n", err)
}
```

## Best Practices

- **Context information**: Add context when adding errors to make debugging easier
- **Check for nil**: The `Add` method safely handles nil errors
- **Validation**: Perfect for form validation and configuration checking
- **Batch processing**: Collect all errors instead of failing on first error
- **Error wrapping**: Can be combined with error wrapping for better error chains

## Use Cases

- **Form validation**: Collect all validation errors for user feedback
- **Configuration validation**: Check multiple config parameters
- **Batch processing**: Report all failures in a batch operation
- **Data migration**: Collect errors from processing multiple records
- **API validation**: Validate multiple request parameters

This package provides a clean way to handle multiple errors without stopping at the first failure, giving users comprehensive feedback about what needs to be fixed.
