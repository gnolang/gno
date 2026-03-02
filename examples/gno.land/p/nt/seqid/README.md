# `seqid` - Sequential ID Generator

A simple sequential ID generator that produces unique identifiers with proper lexicographic ordering for use as AVL tree keys. Provides both binary and human-readable string representations.

## Features

- **Sequential generation**: Generates unique, incrementing IDs
- **AVL tree compatible**: Keys maintain proper sort order in AVL trees
- **Multiple formats**: Binary and human-readable string representations
- **Overflow protection**: Safe handling of ID overflow scenarios
- **Human-friendly encoding**: Uses Crockford Base32 for readable IDs
- **Lexicographic ordering**: String representations maintain numeric order

## Usage

```go
import "gno.land/p/nt/seqid"

// Create ID generator
var idGen seqid.ID

// Generate next ID
userID := idGen.Next()  // ID(1)
postID := idGen.Next()  // ID(2)

// Use as AVL tree keys
var users avl.Tree
users.Set(userID.String(), &User{Name: "Alice"})
users.Set(postID.String(), &Post{Title: "Hello"})

// Binary format for internal storage
var internalStore avl.Tree
internalStore.Set(userID.Binary(), userData)
```

## ID Formats

### String Format (Human-Readable)
```go
id := seqid.ID(42)
str := id.String()  // "1A" (Crockford Base32)

// Parse back from string
parsed, err := seqid.FromString(str)
if err != nil {
    // Handle parsing error
}
```

### Binary Format (Compact)
```go
id := seqid.ID(42)
binary := id.Binary()  // 8-byte big-endian binary

// Parse back from binary
parsed, ok := seqid.FromBinary(binary)
if !ok {
    // Handle parsing error
}
```

## API

```go
type ID uint64

// ID generation
func (i *ID) Next() ID              // Panics on overflow
func (i *ID) TryNext() (ID, bool)   // Returns false on overflow

// String representation (Crockford Base32)
func (i ID) String() string
func FromString(s string) (ID, error)

// Binary representation (8-byte big-endian)
func (i ID) Binary() string
func FromBinary(b string) (ID, bool)
```

## Examples

### User Registration System

```go
type UserRegistry struct {
    users  *avl.Tree
    nextID seqid.ID
}

func NewUserRegistry() *UserRegistry {
    return &UserRegistry{
        users:  avl.NewTree(),
        nextID: seqid.ID(0),
    }
}

func (ur *UserRegistry) RegisterUser(name, email string) string {
    userID := ur.nextID.Next()
    
    user := &User{
        ID:    userID,
        Name:  name,
        Email: email,
    }
    
    // Use string representation as key
    key := userID.String()
    ur.users.Set(key, user)
    
    return key  // Return human-readable ID
}

func (ur *UserRegistry) GetUser(idStr string) (*User, error) {
    // Parse user input (case-insensitive, handles ambiguous chars)
    id, err := seqid.FromString(idStr)
    if err != nil {
        return nil, err
    }
    
    // Always use canonical string representation
    canonicalKey := id.String()
    
    user, exists := ur.users.Get(canonicalKey)
    if !exists {
        return nil, errors.New("user not found")
    }
    
    return user.(*User), nil
}
```

### Blog Post Management

```go
type Blog struct {
    posts  *avl.Tree
    nextID seqid.ID
}

func (b *Blog) CreatePost(title, content string) string {
    postID := b.nextID.Next()
    
    post := &Post{
        ID:        postID,
        Title:     title,
        Content:   content,
        CreatedAt: time.Now(),
    }
    
    // Posts sorted chronologically by ID
    b.posts.Set(postID.String(), post)
    
    return postID.String()  // Return post ID
}

func (b *Blog) GetRecentPosts(limit int) []*Post {
    var posts []*Post
    count := 0
    
    // Iterate in reverse order (newest first)
    b.posts.ReverseIterate("", "", func(key string, value any) bool {
        if count >= limit {
            return true // Stop iteration
        }
        
        posts = append(posts, value.(*Post))
        count++
        return false
    })
    
    return posts
}
```

### Database with Internal Storage

```go
type Database struct {
    records *avl.Tree  // Binary keys for efficiency
    nextID  seqid.ID
}

func (db *Database) Store(data []byte) seqid.ID {
    recordID := db.nextID.Next()
    
    record := &Record{
        ID:   recordID,
        Data: data,
    }
    
    // Use binary representation for internal storage
    db.records.Set(recordID.Binary(), record)
    
    return recordID
}

func (db *Database) Get(id seqid.ID) (*Record, bool) {
    record, exists := db.records.Get(id.Binary())
    if !exists {
        return nil, false
    }
    
    return record.(*Record), true
}
```

### URL Shortener

```go
type URLShortener struct {
    urls   *avl.Tree
    nextID seqid.ID
}

func (us *URLShortener) ShortenURL(longURL string) string {
    urlID := us.nextID.Next()
    
    urlRecord := &URLRecord{
        ID:      urlID,
        LongURL: longURL,
        Hits:    0,
    }
    
    shortCode := urlID.String()  // Human-readable short code
    us.urls.Set(shortCode, urlRecord)
    
    return shortCode
}

func (us *URLShortener) ResolveURL(shortCode string) (string, error) {
    // Handle user input variations (case, ambiguous chars)
    id, err := seqid.FromString(shortCode)
    if err != nil {
        return "", err
    }
    
    // Use canonical representation
    canonicalCode := id.String()
    
    record, exists := us.urls.Get(canonicalCode)
    if !exists {
        return "", errors.New("URL not found")
    }
    
    urlRecord := record.(*URLRecord)
    urlRecord.Hits++
    
    return urlRecord.LongURL, nil
}
```

## Format Characteristics

### String Format (Crockford Base32)
- **Length**: 7 characters for IDs [0, 2^34), 13 characters for larger values
- **Character set**: 0123456789ABCDEFGHJKMNPQRSTVWXYZ
- **Case insensitive**: Accepts both upper and lowercase
- **Error resistant**: I/L/1 and O/0 are interchangeable
- **Lexicographic ordering**: String comparison matches numeric ordering

### Binary Format
- **Length**: Always 8 bytes (64-bit big-endian)
- **Compact**: Most space-efficient representation
- **Ordering**: Byte comparison matches numeric ordering
- **Use case**: Internal storage, network protocols

## Overflow Handling

```go
// Safe ID generation with overflow check
func generateIDSafely(idGen *seqid.ID) (seqid.ID, error) {
    newID, ok := idGen.TryNext()
    if !ok {
        return 0, errors.New("ID generator overflow")
    }
    return newID, nil
}

// Maximum ID value
const maxID = seqid.ID(1<<64 - 1)

// Check if near overflow
if currentID > maxID-1000 {
    // Consider implementing ID recycling or sharding
}
```

## Best Practices

- **User input**: Always parse with `FromString()` then use `String()` for canonical keys
- **Internal storage**: Use `Binary()` for space efficiency
- **Public APIs**: Use `String()` for human-readable identifiers
- **Ordering**: Both formats maintain proper lexicographic ordering
- **Overflow**: Use `TryNext()` for critical applications

## Use Cases

- **User registration**: Human-friendly user IDs
- **URL shortening**: Short, readable codes
- **Post/comment IDs**: Chronologically ordered content
- **Database keys**: Efficient sequential keys
- **API endpoints**: Clean, ordered resource identifiers
- **File naming**: Sequential file naming with proper sorting

## Dependencies

- `gno.land/p/nt/cford32` - Crockford Base32 encoding
- `encoding/binary` - Binary encoding utilities

This package provides a robust foundation for generating sequential identifiers that work seamlessly with AVL trees and provide both human-readable and binary representations.
