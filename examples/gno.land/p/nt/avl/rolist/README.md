# `avl/rolist` - Read-Only List with Safe Value Transformation

A read-only wrapper for `avl/list` that provides safe value transformation, allowing you to expose filtered or transformed views of sensitive data without modification capabilities.

## Features

- **Read-only access**: Prevents modification of the underlying list
- **Value transformation**: Apply transformations when accessing values
- **Data sanitization**: Filter out sensitive fields when exposing data
- **Safe sharing**: Share data views without risking corruption
- **Performance optimized**: Transformations applied only when accessing data

## Usage

```go
import (
    "gno.land/p/nt/avl/list"
    "gno.land/p/nt/avl/rolist"
)

// Original data structure with sensitive fields
type User struct {
    Name     string
    Balance  int
    Internal string // sensitive field
}

// Create and populate list
var privateList list.List
privateList.Append(&User{
    Name:     "Alice", 
    Balance:  100,
    Internal: "sensitive_data",
})

// Create transformation function to remove sensitive data
makeEntrySafeFn := func(v any) any {
    u := v.(*User)
    return &User{
        Name:     u.Name,
        Balance:  u.Balance,
        Internal: "", // omit sensitive data
    }
}

// Create read-only list with transformation
readOnlyList := rolist.NewReadOnlyList(&privateList, makeEntrySafeFn)

// Access safely transformed data
safeUser := readOnlyList.Get(0).(*User)
// safeUser.Internal is now ""
```

## API

```go
type ReadOnlyList struct {
    // private fields
}

// Constructor
func NewReadOnlyList(list list.IList, makeEntrySafeFn func(any) any) *ReadOnlyList

// Read-only operations
func (rol *ReadOnlyList) Get(index int) any
func (rol *ReadOnlyList) Size() int
func (rol *ReadOnlyList) IsEmpty() bool
func (rol *ReadOnlyList) IndexOf(value any) int

// Iteration
func (rol *ReadOnlyList) ForEach(fn func(int, any) bool)
func (rol *ReadOnlyList) ForEachInRange(start, end int, fn func(int, any) bool)
```

## Examples

### Public API for Private Data

```go
type BankAccount struct {
    ID          string
    PublicName  string
    Balance     int64
    AccountKey  string  // sensitive
    InternalID  int     // sensitive
}

type PublicBankAccount struct {
    ID         string
    PublicName string  
    Balance    int64
    // No sensitive fields
}

type BankingSystem struct {
    accounts *list.List
    publicView *rolist.ReadOnlyList
}

func NewBankingSystem() *BankingSystem {
    accounts := list.New()
    
    // Create public view with transformation
    publicView := rolist.NewReadOnlyList(accounts, func(v any) any {
        account := v.(*BankAccount)
        return &PublicBankAccount{
            ID:         account.ID,
            PublicName: account.PublicName,
            Balance:    account.Balance,
        }
    })
    
    return &BankingSystem{
        accounts:   accounts,
        publicView: publicView,
    }
}

func (bs *BankingSystem) AddAccount(account *BankAccount) {
    bs.accounts.Append(account)
}

// Safe public access
func (bs *BankingSystem) GetPublicAccounts() *rolist.ReadOnlyList {
    return bs.publicView
}

// Internal access (full data)
func (bs *BankingSystem) getInternalAccount(index int) *BankAccount {
    return bs.accounts.Get(index).(*BankAccount)
}
```

### User Profile System

```go
type UserProfile struct {
    Username    string
    Email       string
    DisplayName string
    PrivateNotes string  // admin only
    Password    string   // never expose
    AdminFlags  int      // internal
}

type PublicProfile struct {
    Username    string
    DisplayName string
    // Only safe fields
}

func CreateUserProfileSystem() {
    profiles := list.New()
    
    // Add some users
    profiles.Append(&UserProfile{
        Username:     "alice",
        Email:        "alice@example.com", 
        DisplayName:  "Alice Smith",
        PrivateNotes: "VIP customer",
        Password:     "hashed_password",
        AdminFlags:   42,
    })
    
    // Create public read-only view
    publicProfiles := rolist.NewReadOnlyList(profiles, func(v any) any {
        profile := v.(*UserProfile)
        return &PublicProfile{
            Username:    profile.Username,
            DisplayName: profile.DisplayName,
        }
    })
    
    // Public API
    publicProfiles.ForEach(func(index int, value any) bool {
        profile := value.(*PublicProfile)
        fmt.Printf("User: %s (%s)\n", profile.Username, profile.DisplayName)
        return false
    })
}
```

### Audit Log with Redaction

```go
type AuditEntry struct {
    Timestamp   time.Time
    UserID      string
    Action      string
    Details     string
    IPAddress   string  // sensitive
    SessionKey  string  // sensitive
}

type PublicAuditEntry struct {
    Timestamp time.Time
    Action    string
    Details   string
    // IP and session info redacted
}

func CreateAuditSystem() {
    auditLog := list.New()
    
    // Create redacted view for public consumption
    publicAuditLog := rolist.NewReadOnlyList(auditLog, func(v any) any {
        entry := v.(*AuditEntry)
        return &PublicAuditEntry{
            Timestamp: entry.Timestamp,
            Action:    entry.Action, 
            Details:   redactSensitiveInfo(entry.Details),
        }
    })
    
    return publicAuditLog
}

func redactSensitiveInfo(details string) string {
    // Remove sensitive information from details
    // Implementation depends on your needs
    return details
}
```

### Configuration with Secret Masking

```go
type AppConfig struct {
    Name        string
    Version     string
    DatabaseURL string  // sensitive
    APIKey      string  // sensitive
    DebugMode   bool
}

type PublicConfig struct {
    Name      string
    Version   string  
    DebugMode bool
    // Secrets masked
}

func ExposePublicConfig(configs *list.List) *rolist.ReadOnlyList {
    return rolist.NewReadOnlyList(configs, func(v any) any {
        config := v.(*AppConfig)
        return &PublicConfig{
            Name:      config.Name,
            Version:   config.Version,
            DebugMode: config.DebugMode,
        }
    })
}
```

## Transformation Functions

### Simple Field Filtering
```go
func removeSecrets(v any) any {
    user := v.(*User)
    return &User{
        Name:    user.Name,
        Balance: user.Balance,
        // Internal field omitted
    }
}
```

### Data Masking
```go
func maskPersonalInfo(v any) any {
    user := v.(*User)
    return &User{
        Name:    maskName(user.Name),      // "Alice" -> "A***e"
        Email:   maskEmail(user.Email),    // "alice@example.com" -> "a***e@e***.com"
        Balance: user.Balance,
    }
}
```

### Role-Based Filtering
```go
func createRoleBasedTransform(userRole string) func(any) any {
    return func(v any) any {
        data := v.(*SensitiveData)
        
        switch userRole {
        case "admin":
            return data // Full access
        case "user":
            return filterUserFields(data)
        case "guest":
            return filterGuestFields(data)
        default:
            return nil // No access
        }
    }
}
```

## Use Cases

- **API security**: Expose safe views of internal data
- **Privacy compliance**: Filter personal information
- **Role-based access**: Show different views based on user roles
- **Audit trails**: Provide redacted logs for compliance
- **Configuration**: Mask secrets in config displays
- **Data sharing**: Share data safely between components

This package is essential for maintaining data security while providing convenient read-only access to transformed or filtered data views.
