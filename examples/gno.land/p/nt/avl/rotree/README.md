# `avl/rotree` - Read-Only Tree with Safe Value Transformation

A read-only wrapper for `avl.Tree` that provides safe value transformation, allowing you to expose filtered or transformed views of sensitive data stored in AVL trees without modification capabilities.

## Features

- **Read-only access**: Prevents modification of the underlying tree
- **Value transformation**: Apply transformations when accessing values
- **Data sanitization**: Filter out sensitive fields when exposing data
- **Safe sharing**: Share tree data views without risking corruption
- **Performance optimized**: Transformations applied only when accessing data
- **Full tree operations**: Support for all read operations including iteration

## Usage

```go
import (
    "gno.land/p/nt/avl"
    "gno.land/p/nt/avl/rotree"
)

// Original data structure with sensitive fields
type User struct {
    Name     string
    Balance  int
    Internal string // sensitive field
}

// Create and populate tree
privateTree := avl.NewTree()
privateTree.Set("alice", &User{
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

// Create read-only tree with transformation
readOnlyTree := rotree.NewReadOnlyTree(privateTree, makeEntrySafeFn)

// Access safely transformed data
safeUser, exists := readOnlyTree.Get("alice")
// safeUser.(*User).Internal is now ""
```

## API

```go
type ReadOnlyTree struct {
    // private fields
}

// Constructor
func NewReadOnlyTree(tree avl.ITree, makeEntrySafeFn func(any) any) *ReadOnlyTree

// Read-only operations  
func (rot *ReadOnlyTree) Size() int
func (rot *ReadOnlyTree) Has(key string) bool
func (rot *ReadOnlyTree) Get(key string) (value any, exists bool)
func (rot *ReadOnlyTree) GetByIndex(index int) (key string, value any)

// Iteration
func (rot *ReadOnlyTree) Iterate(start, end string, cb IterCbFn) bool
func (rot *ReadOnlyTree) ReverseIterate(start, end string, cb IterCbFn) bool
func (rot *ReadOnlyTree) IterateByOffset(offset int, count int, cb IterCbFn) bool
func (rot *ReadOnlyTree) ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool
```

## Examples

### User Management System

```go
type UserAccount struct {
    ID          string
    Username    string
    Email       string
    Balance     int64
    PasswordHash string  // sensitive
    AdminNotes  string   // sensitive
    CreatedAt   time.Time
}

type PublicUserAccount struct {
    ID        string
    Username  string
    CreatedAt time.Time
    // Sensitive fields omitted
}

type UserManager struct {
    users      *avl.Tree
    publicView *rotree.ReadOnlyTree
}

func NewUserManager() *UserManager {
    users := avl.NewTree()
    
    // Create public view with transformation
    publicView := rotree.NewReadOnlyTree(users, func(v any) any {
        account := v.(*UserAccount)
        return &PublicUserAccount{
            ID:        account.ID,
            Username:  account.Username,
            CreatedAt: account.CreatedAt,
        }
    })
    
    return &UserManager{
        users:      users,
        publicView: publicView,
    }
}

func (um *UserManager) AddUser(account *UserAccount) {
    um.users.Set(account.ID, account)
}

// Safe public access
func (um *UserManager) GetPublicUsers() *rotree.ReadOnlyTree {
    return um.publicView
}

// Internal access (full data)
func (um *UserManager) getInternalUser(id string) *UserAccount {
    user, exists := um.users.Get(id)
    if !exists {
        return nil
    }
    return user.(*UserAccount)
}
```

### Configuration Management

```go
type AppConfig struct {
    ServiceName string
    Version     string
    Port        int
    DatabaseURL string  // sensitive
    APISecret   string  // sensitive
    JWTKey      string  // sensitive
    Debug       bool
}

type PublicConfig struct {
    ServiceName string
    Version     string
    Port        int
    Debug       bool
    // Secrets omitted
}

func CreateConfigSystem() {
    configs := avl.NewTree()
    
    // Add configuration
    configs.Set("app", &AppConfig{
        ServiceName: "MyApp",
        Version:     "1.0.0",
        Port:        8080,
        DatabaseURL: "postgres://secret",
        APISecret:   "super-secret-key",
        JWTKey:      "jwt-signing-key",
        Debug:       true,
    })
    
    // Create public view
    publicConfigs := rotree.NewReadOnlyTree(configs, func(v any) any {
        config := v.(*AppConfig)
        return &PublicConfig{
            ServiceName: config.ServiceName,
            Version:     config.Version,
            Port:        config.Port,
            Debug:       config.Debug,
        }
    })
    
    // Safe to expose publicly
    return publicConfigs
}
```

### Product Catalog with Pricing

```go
type Product struct {
    ID          string
    Name        string
    Description string
    Cost        float64  // internal cost (sensitive)
    Price       float64  // public price
    Inventory   int
    SupplierID  string   // sensitive
}

type PublicProduct struct {
    ID          string
    Name        string
    Description string
    Price       float64
    Available   bool     // calculated from inventory
}

func CreateProductCatalog() *rotree.ReadOnlyTree {
    products := avl.NewTree()
    
    // Add products
    products.Set("prod-1", &Product{
        ID:          "prod-1",
        Name:        "Widget",
        Description: "A useful widget",
        Cost:        10.50,  // don't expose
        Price:       19.99,
        Inventory:   100,
        SupplierID:  "supplier-123",  // don't expose
    })
    
    // Create public catalog
    return rotree.NewReadOnlyTree(products, func(v any) any {
        product := v.(*Product)
        return &PublicProduct{
            ID:          product.ID,
            Name:        product.Name,
            Description: product.Description,
            Price:       product.Price,
            Available:   product.Inventory > 0,
        }
    })
}
```

### Session Management

```go
type Session struct {
    ID          string
    UserID      string
    CreatedAt   time.Time
    LastAccess  time.Time
    IPAddress   string    // sensitive
    UserAgent   string    // sensitive
    InternalKey string    // sensitive
    IsActive    bool
}

type SessionInfo struct {
    ID         string
    UserID     string
    CreatedAt  time.Time
    LastAccess time.Time
    IsActive   bool
    // Sensitive info omitted
}

func CreateSessionManager() {
    sessions := avl.NewTree()
    
    // Create safe view for logging/monitoring
    sessionView := rotree.NewReadOnlyTree(sessions, func(v any) any {
        session := v.(*Session)
        return &SessionInfo{
            ID:         session.ID,
            UserID:     session.UserID,
            CreatedAt:  session.CreatedAt,
            LastAccess: session.LastAccess,
            IsActive:   session.IsActive,
        }
    })
    
    return sessionView
}
```

### Role-Based Access Control

```go
func CreateRoleBasedView(data *avl.Tree, userRole string) *rotree.ReadOnlyTree {
    return rotree.NewReadOnlyTree(data, func(v any) any {
        record := v.(*DataRecord)
        
        switch userRole {
        case "admin":
            return record // Full access
            
        case "manager":
            return &DataRecord{
                ID:      record.ID,
                Title:   record.Title,
                Content: record.Content,
                // Hide sensitive admin fields
            }
            
        case "user":
            return &PublicDataRecord{
                ID:    record.ID,
                Title: record.Title,
                // Very limited view
            }
            
        default:
            return nil // No access
        }
    })
}
```

## Iteration with Transformation

```go
// Iterate over transformed values
publicTree.Iterate("", "", func(key string, value any) bool {
    safeValue := value.(*PublicType)
    fmt.Printf("Key: %s, Value: %+v\n", key, safeValue)
    return false // continue
})

// Count public entries
count := 0
publicTree.Iterate("", "", func(key string, value any) bool {
    count++
    return false
})
```

## Use Cases

- **API security**: Expose safe views of internal data structures
- **Privacy compliance**: Filter personal/sensitive information
- **Role-based access**: Show different views based on user permissions
- **Configuration**: Mask secrets in config displays
- **Audit trails**: Provide redacted data for compliance
- **Data sharing**: Safely share data between components or services
- **Public APIs**: Expose filtered versions of internal data

This package is essential for maintaining data security while providing convenient read-only access to transformed or filtered AVL tree data, perfect for building secure APIs and data sharing systems.
