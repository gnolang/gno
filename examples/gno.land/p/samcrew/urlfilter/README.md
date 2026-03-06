# `urlfilter` - URL-based filtering

Filter items using URL query parameters with toggleable markdown links. Works with AVL tree structures where each filter contains its associated items.

Given filters `["T1", "T2", "size:XL"]` and URL `/shop?filter=T1,size:XL`, it generates toggle links:

- **T1** _(active, click to remove)_
- ~~T2~~ _(inactive, click to add)_  
- **size:XL** _(active, click to remove)_

**Markdown output:**
```markdown
[**T1**](/p/samcrew/urlfilter?filter=size:XL) - [~~T2~~](/p/samcrew/urlfilter=T1,T2,size:XL) - [**size:XL**](/p/samcrew/urlfilter?filter=T1)
```

**Rendered as:**
[**T1**](/p/samcrew/urlfilter?filter=size:XL) - [~~T2~~](/p/samcrew/urlfilter=T1,T2,size:XL) - [**size:XL**](/p/samcrew/urlfilter?filter=T1)

## Usage

The package expects a two-level AVL tree structure:
- **Top level**: Filter names as keys (e.g., "T1", "size:XL", "on_sale")  
- **Second level**: Item trees containing the actual items for each filter

```go
// Build the main filters tree
filters := avl.NewTree()

// Subtree for filter "T1" 
t1Items := avl.NewTree()
t1Items.Set("key1", "item1")
t1Items.Set("key2", "item2")
filters.Set("T1", t1Items)

// Subtree for filter "size:XL"
t2Items := avl.NewTree()
t2Items.Set("key3", "item3")
filters.Set("T2", t2Items)

// URL with active filter "T1"
u, _ := url.Parse("/shop?filter=T1")

// Apply filtering
mdLinks, filteredItems := urlfilter.ApplyFilters(u, filters, "filter") // "filter" for /shop?*filter*=T1

// mdLinks    → Markdown links for toggling filters  
// filteredItems → AVL tree containing only filtered items
```

## API

```go
func ApplyFilters(u *url.URL, items *avl.Tree, paramName string) (string, *avl.Tree)
```

**Parameters:**
- `u`: URL containing query parameters
- `items`: Two-level AVL tree (filters → item trees)
- `paramName`: Query parameter name (e.g., "filter" for /shop?filter=T1)

**URL Format:**
- Single filter: `?filter=T1`
- Multiple filters: `?filter=T1,size:XL,on_sale`
- Filter names are comma-separated

**Returns:**
- **Markdown links**: Toggleable filter links with formatting
- **Filtered items**: AVL tree containing items from active filters
  - If no filters active: returns all items
  - Item keys are preserved, values show which filter matched

# Example

- [/r/gov/dao/v3/memberstore:members?filter=T1](/r/gov/dao/v3/memberstore:members?filter=T1)