# `ownable/exts/authorizable` - Extended Authorization System

An extension of the `ownable` package that adds a secondary authorization layer. Provides a superuser (owner) plus additional authorized users who can perform specific operations.

## Features

- **Two-tier authorization**: Owner (superuser) + authorized users (moderators/admins)
- **Flexible permissions**: Different permission levels for different operations
- **AVL tree backed**: Efficient authorized user management
- **Owner privileges**: Owner can manage authorized users
- **Built on ownable**: Inherits all ownable functionality

## Usage

```go
import "gno.land/p/nt/ownable/exts/authorizable"

// Create authorizable instance
auth := authorizable.NewAuthorizable()

// Owner operations (highest privilege)
auth.RequireOwner() // Only owner can do this

// Add authorized users (owner only)
moderatorAddr := std.Address("g1abc123...")
auth.AddAuthorized(moderatorAddr)

// Check authorization levels
isOwner := auth.CallerIsOwner()
isAuthorized := auth.CallerIsAuthorized() // includes owner
isModerator := auth.CallerIsAuthorized() && !auth.CallerIsOwner()

// Require authorization (owner OR authorized user)
auth.RequireAuthorized()

// Remove authorized user (owner only)
auth.RemoveAuthorized(moderatorAddr)
```

## API

```go
type Authorizable struct {
    *ownable.Ownable  // Embedded ownable functionality
    // private fields
}

// Constructor
func NewAuthorizable() *Authorizable
func NewAuthorizableWithOrigin() *Authorizable

// Authorization management (owner only)
func (a *Authorizable) AddAuthorized(addr std.Address) error
func (a *Authorizable) RemoveAuthorized(addr std.Address) error

// Authorization checks
func (a *Authorizable) CallerIsAuthorized() bool
func (a *Authorizable) IsAuthorized(addr std.Address) bool
func (a *Authorizable) RequireAuthorized()

// User management
func (a *Authorizable) GetAuthorized() []std.Address
func (a *Authorizable) IsOnlyOwnerAuthorized() bool
```

## Examples

### Forum Moderation System

```go
type Forum struct {
    *authorizable.Authorizable
    posts   *avl.Tree
    nextID  int
}

func NewForum() *Forum {
    return &Forum{
        Authorizable: authorizable.NewAuthorizable(),
        posts:        avl.NewTree(),
        nextID:       1,
    }
}

// Anyone can create posts
func (f *Forum) CreatePost(title, content string) int {
    postID := f.nextID
    f.nextID++
    
    post := &Post{
        ID:      postID,
        Title:   title,
        Content: content, 
        Author:  std.CurrentCaller(),
    }
    
    f.posts.Set(fmt.Sprintf("%d", postID), post)
    return postID
}

// Only authorized users (moderators) can delete posts
func (f *Forum) DeletePost(postID int) {
    f.RequireAuthorized() // Moderators or owner
    
    key := fmt.Sprintf("%d", postID)
    f.posts.Remove(key)
}

// Only owner can add/remove moderators
func (f *Forum) AddModerator(addr std.Address) error {
    f.RequireOwner() // Only owner
    return f.AddAuthorized(addr)
}

func (f *Forum) RemoveModerator(addr std.Address) error {
    f.RequireOwner() // Only owner
    return f.RemoveAuthorized(addr)
}
```

### DAO with Admin Roles

```go
type DAO struct {
    *authorizable.Authorizable
    proposals *avl.Tree
    treasury  int64
}

func NewDAO() *DAO {
    return &DAO{
        Authorizable: authorizable.NewAuthorizable(),
        proposals:    avl.NewTree(),
        treasury:     0,
    }
}

// Anyone can create proposals
func (d *DAO) CreateProposal(title, description string) {
    // Proposal creation logic
}

// Authorized users can approve proposals
func (d *DAO) ApproveProposal(proposalID string) {
    d.RequireAuthorized() // Admins or owner
    
    // Approval logic
}

// Only owner can manage treasury funds
func (d *DAO) TransferFunds(to std.Address, amount int64) {
    d.RequireOwner() // Only owner
    
    if amount > d.treasury {
        panic("insufficient funds")
    }
    
    // Transfer logic
    d.treasury -= amount
}

// Owner can delegate admin privileges
func (d *DAO) AddAdmin(addr std.Address) error {
    d.RequireOwner()
    return d.AddAuthorized(addr)
}
```

### Content Management System

```go
type CMS struct {
    *authorizable.Authorizable
    articles *avl.Tree
}

func NewCMS() *CMS {
    return &CMS{
        Authorizable: authorizable.NewAuthorizable(),
        articles:     avl.NewTree(),
    }
}

// Authorized users can publish articles
func (c *CMS) PublishArticle(title, content string) {
    c.RequireAuthorized() // Authors or owner
    
    article := &Article{
        Title:     title,
        Content:   content,
        Author:    std.CurrentCaller(),
        Published: time.Now(),
    }
    
    c.articles.Set(title, article)
}

// Anyone can read articles
func (c *CMS) GetArticle(title string) *Article {
    article, exists := c.articles.Get(title)
    if !exists {
        return nil
    }
    return article.(*Article)
}

// Owner can manage authors
func (c *CMS) AddAuthor(addr std.Address) error {
    c.RequireOwner()
    return c.AddAuthorized(addr)
}

func (c *CMS) RemoveAuthor(addr std.Address) error {
    c.RequireOwner()
    return c.RemoveAuthorized(addr)
}

// List all authors (public info)
func (c *CMS) GetAuthors() []std.Address {
    return c.GetAuthorized()
}
```

### Multi-Signature Wallet

```go
type MultiSigWallet struct {
    *authorizable.Authorizable
    balance    int64
    threshold  int // required signatures
}

func NewMultiSigWallet(signers []std.Address, threshold int) *MultiSigWallet {
    wallet := &MultiSigWallet{
        Authorizable: authorizable.NewAuthorizable(),
        balance:      0,
        threshold:    threshold,
    }
    
    // Add signers as authorized users
    for _, signer := range signers {
        wallet.AddAuthorized(signer)
    }
    
    return wallet
}

// Only authorized signers can propose transactions
func (w *MultiSigWallet) ProposeTransaction(to std.Address, amount int64) {
    w.RequireAuthorized() // Must be a signer
    
    // Create transaction proposal
    // Implementation depends on your multi-sig logic
}

// Owner can add new signers
func (w *MultiSigWallet) AddSigner(addr std.Address) error {
    w.RequireOwner()
    return w.AddAuthorized(addr)
}

// Check if enough signers
func (w *MultiSigWallet) HasEnoughSigners() bool {
    signers := w.GetAuthorized()
    return len(signers) >= w.threshold
}
```

## Permission Patterns

### Three-Tier System
```go
// Owner only
func (c *Contract) OwnerOnlyFunction() {
    c.RequireOwner()
    // Owner-exclusive logic
}

// Authorized users (including owner)
func (c *Contract) AuthorizedFunction() {
    c.RequireAuthorized()
    // Moderator/admin logic
}

// Public function
func (c *Contract) PublicFunction() {
    // Anyone can call
}
```

### Role-Based Checks
```go
func (c *Contract) HandleAction(action string) {
    caller := std.CurrentCaller()
    
    if c.CallerIsOwner() {
        // Owner can do anything
        c.executeAction(action)
    } else if c.IsAuthorized(caller) {
        // Authorized users have limited actions
        if action == "moderate" || action == "approve" {
            c.executeAction(action)
        } else {
            panic("unauthorized action")
        }
    } else {
        panic("not authorized")
    }
}
```

## Error Handling

The package defines specific errors for authorization failures:

```go
var (
    ErrAlreadyAuthorized = errors.New("address is already authorized")
    ErrNotAuthorized     = errors.New("address is not authorized")
    ErrCannotRemoveOwner = errors.New("cannot remove owner from authorized list")
)
```

## Use Cases

- **Forum moderation**: Owner + moderators structure
- **DAO governance**: Founder + admin council
- **Content management**: Editor-in-chief + writers
- **Multi-signature wallets**: Key holders + recovery key
- **Gaming guilds**: Guild master + officers
- **Subscription services**: Admin + customer support

This package provides a clean way to implement hierarchical permission systems where you need both a superuser and a secondary authorization layer.
