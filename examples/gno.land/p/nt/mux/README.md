# `mux` - HTTP-style Router for Gno

A routing and rendering library that provides HTTP ServeMux-like functionality for handling dynamic path-based requests in Gno contracts. Perfect for implementing web-like interfaces in your Gno applications.

## Features

- **Dynamic routing**: Support for path variables like `/users/{id}`
- **Handler functions**: Associate routes with specific handler functions  
- **Request/Response**: Familiar HTTP-style request and response objects
- **Path parsing**: Extract variables from URL paths
- **Pattern matching**: Flexible route pattern matching

## Usage

```go
import "gno.land/p/nt/mux"

// Create a new router
router := mux.NewRouter()

// Register routes with handlers
router.HandleFunc("hello/{name}", func(res *mux.ResponseWriter, req *mux.Request) {
    name := req.GetVar("name")
    if name != "" {
        res.Write("Hello, " + name + "!")
    } else {
        res.Write("Hello, world!")
    }
})

router.HandleFunc("users/{id}/posts/{postId}", func(res *mux.ResponseWriter, req *mux.Request) {
    userID := req.GetVar("id")
    postID := req.GetVar("postId")
    res.Write("User " + userID + ", Post " + postID)
})

// Handle requests
output := router.Render("/hello/Alice")  // "Hello, Alice!"
output = router.Render("/users/123/posts/456")  // "User 123, Post 456"
```

## Route Patterns

Routes can include dynamic segments enclosed in braces:

- `hello/{name}` - Matches `/hello/Alice`, `/hello/Bob`
- `users/{id}` - Matches `/users/123`, `/users/abc`  
- `api/v1/users/{id}/posts/{postId}` - Multiple variables
- `static/path` - Exact match only

## API

### Router
```go
type Router struct {
    // private fields
}

func NewRouter() *Router
func (r *Router) HandleFunc(pattern string, handler func(*ResponseWriter, *Request))
func (r *Router) Handle(pattern string, handler Handler)
func (r *Router) Render(path string) string
```

### Request
```go
type Request struct {
    // private fields
}

func (r *Request) GetVar(key string) string
```

### Response Writer
```go
type ResponseWriter struct {
    // private fields
}

func (w *ResponseWriter) Write(data string)
```

### Handler Interface
```go
type Handler interface {
    ServeGno(res *ResponseWriter, req *Request)
}
```

## Advanced Example

```go
// Blog-style routing
router := mux.NewRouter()

// Home page
router.HandleFunc("", func(res *mux.ResponseWriter, req *mux.Request) {
    res.Write("Welcome to the blog!")
})

// Blog post by ID
router.HandleFunc("posts/{id}", func(res *mux.ResponseWriter, req *mux.Request) {
    postID := req.GetVar("id")
    post := getPost(postID)
    if post != nil {
        res.Write("Post: " + post.Title)
    } else {
        res.Write("Post not found")
    }
})

// User profile
router.HandleFunc("user/{username}", func(res *mux.ResponseWriter, req *mux.Request) {
    username := req.GetVar("username")
    profile := getUserProfile(username)
    res.Write("Profile: " + profile.Name)
})

// Category with pagination
router.HandleFunc("category/{cat}/page/{page}", func(res *mux.ResponseWriter, req *mux.Request) {
    category := req.GetVar("cat")
    page := req.GetVar("page")
    posts := getPostsByCategory(category, page)
    res.Write("Category " + category + ", Page " + page)
})
```

## Integration with Gno Contracts

```go
var router *mux.Router

func init() {
    router = mux.NewRouter()
    setupRoutes()
}

func Render(path string) string {
    return router.Render(path)
}

func setupRoutes() {
    router.HandleFunc("", homePage)
    router.HandleFunc("user/{id}", userProfile)
    router.HandleFunc("posts/{category}", categoryPosts)
}
```

This package enables you to create sophisticated web-like interfaces for your Gno contracts, making them more user-friendly and easier to navigate.
