# `avl/pager` - Pagination for AVL Trees

A pagination utility for AVL trees that provides easy navigation through large datasets. Supports URL-based pagination with customizable page size and navigation.

## Features

- **URL-based pagination**: Parse page parameters from URL query strings
- **Configurable page size**: Set default and custom page sizes
- **Navigation helpers**: Easy previous/next page navigation
- **Total count tracking**: Know total items and pages
- **Reverse pagination**: Support for reverse chronological order
- **Item extraction**: Convert tree nodes to structured items

## Usage

```go
import (
    "gno.land/p/nt/avl/pager"
    "gno.land/p/nt/avl/rotree"
)

// Create read-only tree wrapper
tree := avl.NewTree()
// ... populate tree ...
readOnlyTree := rotree.NewReadOnlyTree(tree)

// Create pager
p := &pager.Pager{
    Tree:            readOnlyTree,
    PageQueryParam:  "page",     // URL param name for page number
    SizeQueryParam:  "size",     // URL param name for page size
    DefaultPageSize: 10,         // Default items per page
    Reversed:        false,      // Normal order (set true for reverse)
}

// Parse URL and get page
u, _ := url.Parse("/posts?page=2&size=20")
page := p.MustGetPageByURL(u)

// Access page information
fmt.Printf("Page %d of %d\n", page.PageNumber, page.TotalPages)
fmt.Printf("Items %d-%d of %d total\n", 
    (page.PageNumber-1)*page.PageSize + 1,
    min(page.PageNumber*page.PageSize, page.TotalItems),
    page.TotalItems)
```

## API

### Pager Configuration
```go
type Pager struct {
    Tree            rotree.IReadOnlyTree
    PageQueryParam  string  // Default: "page"
    SizeQueryParam  string  // Default: "size"  
    DefaultPageSize int     // Default: 10
    Reversed        bool    // Default: false (forward order)
}
```

### Page Information
```go
type Page struct {
    Items      []Item  // Current page items
    PageNumber int     // Current page number (1-based)
    PageSize   int     // Items per page
    TotalItems int     // Total items in dataset
    TotalPages int     // Total number of pages
    HasPrev    bool    // True if previous page exists
    HasNext    bool    // True if next page exists
    Pager      *Pager  // Reference to parent pager
}
```

### Methods
```go
// Get page from URL
func (p *Pager) GetPageByURL(u *url.URL) (*Page, error)
func (p *Pager) MustGetPageByURL(u *url.URL) *Page

// Get specific page
func (p *Pager) GetPage(pageNum, pageSize int) (*Page, error)

// Navigation URLs
func (page *Page) PrevURL(baseURL string) string
func (page *Page) NextURL(baseURL string) string
func (page *Page) PageURL(baseURL string, pageNum int) string
```

## Examples

### Blog Post Pagination

```go
// Setup
posts := avl.NewTree()
// ... add blog posts ...

pager := &pager.Pager{
    Tree:            rotree.NewReadOnlyTree(posts),
    PageQueryParam:  "page",
    SizeQueryParam:  "size", 
    DefaultPageSize: 5,
    Reversed:        true,  // Newest first
}

// In your Render function
func Render(path string) string {
    u, _ := url.Parse(path)
    page := pager.MustGetPageByURL(u)
    
    var result strings.Builder
    result.WriteString(fmt.Sprintf("# Blog Posts (Page %d)\n\n", page.PageNumber))
    
    // Render posts
    for _, item := range page.Items {
        post := item.Value.(*BlogPost)
        result.WriteString(fmt.Sprintf("## %s\n%s\n\n", post.Title, post.Summary))
    }
    
    // Navigation
    result.WriteString("---\n")
    if page.HasPrev {
        result.WriteString(fmt.Sprintf("[Previous](%s) | ", page.PrevURL("/blog")))
    }
    result.WriteString(fmt.Sprintf("Page %d of %d", page.PageNumber, page.TotalPages))
    if page.HasNext {
        result.WriteString(fmt.Sprintf(" | [Next](%s)", page.NextURL("/blog")))
    }
    
    return result.String()
}
```

### User Directory

```go
type UserDirectory struct {
    users  *avl.Tree
    pager  *pager.Pager
}

func NewUserDirectory() *UserDirectory {
    users := avl.NewTree()
    
    return &UserDirectory{
        users: users,
        pager: &pager.Pager{
            Tree:            rotree.NewReadOnlyTree(users),
            PageQueryParam:  "p",
            SizeQueryParam:  "n",
            DefaultPageSize: 20,
        },
    }
}

func (ud *UserDirectory) RenderPage(path string) string {
    u, _ := url.Parse(path)
    page := ud.pager.MustGetPageByURL(u)
    
    var html strings.Builder
    html.WriteString("<div class='user-directory'>\n")
    html.WriteString(fmt.Sprintf("<h2>Users (%d total)</h2>\n", page.TotalItems))
    
    html.WriteString("<ul>\n")
    for _, item := range page.Items {
        user := item.Value.(*User)
        html.WriteString(fmt.Sprintf("<li>%s - %s</li>\n", user.Name, user.Email))
    }
    html.WriteString("</ul>\n")
    
    // Pagination controls
    html.WriteString("<div class='pagination'>\n")
    for i := 1; i <= page.TotalPages; i++ {
        if i == page.PageNumber {
            html.WriteString(fmt.Sprintf("<strong>%d</strong> ", i))
        } else {
            html.WriteString(fmt.Sprintf("<a href='%s'>%d</a> ", 
                page.PageURL("/users", i), i))
        }
    }
    html.WriteString("</div>\n")
    
    html.WriteString("</div>\n")
    return html.String()
}
```

### API Response Pagination

```go
func GetAPIResponse(path string) map[string]interface{} {
    u, _ := url.Parse(path)
    page := pager.MustGetPageByURL(u)
    
    // Convert items to API format
    items := make([]map[string]interface{}, len(page.Items))
    for i, item := range page.Items {
        items[i] = map[string]interface{}{
            "id":    item.Key,
            "data":  item.Value,
        }
    }
    
    return map[string]interface{}{
        "data": items,
        "pagination": map[string]interface{}{
            "page":        page.PageNumber,
            "size":        page.PageSize,
            "total_items": page.TotalItems,
            "total_pages": page.TotalPages,
            "has_prev":    page.HasPrev,
            "has_next":    page.HasNext,
        },
    }
}
```

## Navigation Helpers

```go
// Generate navigation URLs
baseURL := "/posts"

// Previous page URL
if page.HasPrev {
    prevURL := page.PrevURL(baseURL)
    // Use prevURL for navigation
}

// Next page URL  
if page.HasNext {
    nextURL := page.NextURL(baseURL)
    // Use nextURL for navigation
}

// Specific page URL
pageURL := page.PageURL(baseURL, 5)  // URL for page 5
```

## Use Cases

- **Blog pagination**: Navigate through blog posts
- **User directories**: Browse user lists with pagination
- **Product catalogs**: Paginate through product listings
- **Search results**: Page through search results
- **API responses**: Provide paginated API endpoints
- **Data exploration**: Browse large datasets interactively

This package makes it easy to add pagination to any AVL tree-based data structure, providing both functional pagination logic and URL generation helpers for web interfaces.
