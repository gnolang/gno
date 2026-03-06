# Example `minisocial` dApp

We will create a MiniSocial [realm](../resources/realms.md),
a minimalist social media application. This tutorial will showcase a full local
development flow for Gno, using all the tools covered in previous tutorials.

Find the full app on [this link](https://gno.land/r/docs/minisocial/v1).

## Prerequisites

See [Local Development with gnodev](./local-dev-with-gnodev.md) for setup instructions.

## Setup

Start by creating a folder that will contain your Gno code:

```sh
mkdir minisocial
cd minisocial
```

Next, initialize a `gnomod.toml` file. This file declares the package path of your
realm and is used by Gno tools. Run the following command to create a `gnomod.toml` file:

```sh
gno mod init gno.land/r/example/minisocial
```

In this case, we'll be using the `examples` namespace, but you can change this to
the namespace of your liking later.

Next, in the same folder, start by creating three files:

```sh
touch types.gno minisocial.gno render.gno
```

While all code can be stored in a single file, separating logical units,
such as types, business logic, and rendering can make your realm more readable.

## Core functionality

### `types.gno`

We can use `types.gno` file to store our types and their functionality. We will be
importing some standard library packages, as well as some pure packages directly
from the chain.

First, let's declare a `Post` struct that will hold all the data of a single post.
We import the `time` package, which allows us to handle time-related functionality.

[embedmd]:# (../_assets/minisocial/types-1.gno go)
```go
package minisocial

import (
	"time" // For handling time operations
)

// Post defines the main data we keep about each post
type Post struct {
	text      string    // Main text body
	author    address   // Address of the post author, provided by the execution context
	createdAt time.Time // When the post was created
}
```

The `address` keyword is a built-in keyword type represents a Gno address.

Standard libraries such as `time` are ported over directly from Go. Check out the
[Go-Gno Compatability](../resources/go-gno-compatibility.md) page for more info.

### `posts.gno`

In this file, we will define the main functions for creating, updating, and deleting
posts. Let's start with top level variables - they are the anchor points of our
app, as they are persisted to storage after each transaction:

[embedmd]:# (../_assets/minisocial/posts-0.gno go)
```go
package minisocial

var posts []*Post
```

The `posts` slice will hold our all newly created posts.

Next, in the same file, let's create a function to create new posts. This function
will be [exported](https://go.dev/tour/basics/3), meaning it will be callable via
a transaction by anyone.

[embedmd]:# (../_assets/minisocial/posts-1.gno go /\/\/ CreatePost/ $)
```go
// CreatePost creates a new post
// As the function modifies state (i.e. the `posts` slice),
// it needs to be crossing. This is defined by the first argument being of type `realm`
func CreatePost(_ realm, text string) error {
	// If the body of the post is empty, return an error
	if text == "" {
		return errors.New("empty post text")
	}

	// Append the new post to the list
	posts = append(posts, &Post{
		text:      text,                              // Set the input text
		author:    runtime.PreviousRealm().Address(), // The author of the address is the previous realm, the realm that called this one
		createdAt: time.Now(),                        // Capture the time of the transaction, in this case the block timestamp
	})

	return nil
}
```

A few things to note:
- In Gno, returning errors **_does not_** revert any state changes. Follow Go's
  best practices: return early in your code and modify state only after you are sure all
  security checks in your code have passed. To discard (revert) state changes,
  use `panic()`.
- To get the caller of `CreatePost`, we need to import `chain/runtime`,
which provides access to the function caller, and use `runtime.PreviousRealm.Address()`.
Check out the [realm concept page](../resources/realms.md) & the 
[`chain/runtime` package](../resources/gno-stdlibs.md) reference page for more info.
- In Gno, `time.Now()` returns the timestamp of the block the transaction was
included in, instead of the system time.

:::info Lint & format

The `gno` binary provides tooling which can help you write correct code.
You can use `gno lint` and `gno tool fmt` to lint and format your code,
respectively.
:::

## Rendering

Let's start building the "front end" of our app.

One of the core features of Gno is that developers can simply provide a Markdown
view of their realm state directly in Gno, removing the need for using complex
frontend frameworks, languages, and clients. To learn more about this, check out
[Exploring Gno.land](../users/explore-with-gnoweb.md).

The easiest way to develop this part of our Gno app is to run `gnodev`, which
contains a built-in Gno.land node, a built-in instance of `gnoweb`, fast hot
reload, and automatic balance premining. Using `gnodev` will allow us to see our
code changes live.

Let's start by running `gnodev` inside our `minisocial/` folder:

```
❯ gnodev
Accounts    ┃ I default address imported name=test1 addr=g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
GnoWeb      ┃ I using default package path=gno.land/r/example/minisocial
Proxy       ┃ I lazy loading is enabled. packages will be loaded only upon a request via a query or transaction. loader=local<threads-full>/root<examples>
Node        ┃ I packages paths=[gno.land/r/example/minisocial]
Event       ┃ I sending event to clients clients=0 type=NODE_RESET event={}
GnoWeb      ┃ I gnoweb started lisn=http://127.0.0.1:8888
--- READY   ┃ I for commands and help, press `h` took=1.689893s
```

If we didn't make any errors in our code, we should get the output as presented
above. If not, follow the stack trace and fix any errors that might have showed up.

Next, we can open the `gnoweb` instance by opening the local listener at
[`127.0.0.1:8888`](http://127.0.0.1:8888). `gnodev` is configured to open
the package path you're working on by default.

Since a `Render()` function is not defined yet, `gnoweb` will return an error.
Let's start fixing this, in `render.gno`:

```go
package minisocial

func Render(_ string) string {
    return "# MiniSocial"
}
```

`gnodev` will detect changes in your code and automatically reload, and you
should get `MiniSocial`  rendered as a Header 1 in `gnoweb` 🎉

Let's start by slowly adding more and more functionality:

[embedmd]:# (../_assets/minisocial/render-0.gno go)
```go
package minisocial

func Render(_ string) string {
	output := "# MiniSocial\n\n" // \n is needed just like in standard Markdown

	// Handle the edge case
	if len(posts) == 0 {
		output += "No posts.\n"
		return output
	}

	// Let's append the text of each post to the output
	for _, post := range posts {
		output += post.text + "\n\n"
	}

	return output
}
```

We can now use `gnokey` to call the `CreatePost` function and see how our posts
look rendered on `gnoweb`. Let's use the [Docs] page to obtain the `gnokey` command:

```sh
gnokey maketx call \
-pkgpath "gno.land/r/example/minisocial" \
-func "CreatePost" \
-args "This is my first post" \
-gas-fee 1000000ugnot -gas-wanted 5000000 \
-broadcast \
-chainid "dev" \
-remote "tcp://127.0.0.1:26657" \
{MYKEY}
```

If the transaction went through, we should see `This is my first post` under the
header.

We can make this a bit prettier by introducing a custom `String()` method on
the `Post` struct, in `types.gno`:

[embedmd]:# (../_assets/minisocial/types-2.gno go /\/\/ String/ $)
```go
// String stringifies a Post
func (p Post) String() string {
	out := p.text + "\n\n"

	// We can use `ufmt` to format strings, and the built-in time library formatting function
	out += ufmt.Sprintf("_by %s on %s_, ", p.author, p.createdAt.Format("02 Jan 2006, 15:04"))
	out += "\n\n"

	return out
}
```

Here, package `ufmt` is used to provide string formatting functionality. It can
be imported via with `gno.land/p/nt/ufmt/v0`.

With this, we can expand our `Render()` function in `posts.gno` as follows:

[embedmd]:# (../_assets/minisocial/render-1.gno go)
```go
package minisocial

import "gno.land/p/nt/ufmt/v0" // Gno counterpart to `fmt`, for formatting strings

func Render(_ string) string {
	output := "# MiniSocial\n\n" // \n is needed just like in standard Markdown

	// Handle the edge case
	if len(posts) == 0 {
		output += "No posts.\n"
		return output
	}

	// Let's append the text of each post to the output
	for i, post := range posts {
		// Let's append some post metadata
		output += ufmt.Sprintf("#### Post #%d\n\n", i)
		// Add the stringified post
		output += post.String()
		// Add a line break for cleaner UI
		output += "---\n\n"
	}

	return output
}
```

Now, try publishing a few more posts to see that the rendering works properly.

## Testing our code

Testing is an essential part of developing reliable applications.
Here we will cover a simple test case and then showcase a more advanced approach
using Table-Driven Tests (TDT), a pattern commonly used in Go.

Let's create a `post_test.gno` file, and add the following code:

[embedmd]:# (../_assets/minisocial/posts_test-0.gno go)
```go
package minisocial

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0" // Provides testing utilities
)

func TestCreatePostSingle(t *testing.T) {
	// Get a test address for alice
	aliceAddr := testutils.TestAddress("alice")
	// TestSetRealm sets the realm caller, in this case Alice
	testing.SetRealm(testing.NewUserRealm(aliceAddr))

	text1 := "Hello World!"

	// To call a crossing function, we specify the `cross` keyword
	// This matches the first argument of type realm in the function itself
	err := CreatePost(cross, text1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Get the rendered page
	got := Render("")

	// Content should have the text and alice's address in it
	if !(strings.Contains(got, text1) && strings.Contains(got, aliceAddr.String())) {
		t.Fatal("expected render to contain text & alice's address")
	}
}
```

We can add the following test showcasing how TDT works in Gno:

[embedmd]:# (../_assets/minisocial/posts_test-1.gno go /func TestCreatePostMultiple/ $)
```go
func TestCreatePostMultiple(t *testing.T) {
	// Initialize a slice to hold the test posts and their authors
	posts := []struct {
		text   string
		author string
	}{
		{"Hello World!", "alice"},
		{"This is some new text!", "bob"},
		{"Another post by alice", "alice"},
		{"A post by charlie!", "charlie"},
	}

	for _, p := range posts {
		// Set the appropriate caller realm based on the author
		authorAddr := testutils.TestAddress(p.author)
		testing.SetRealm(testing.NewUserRealm(authorAddr))

		// Create the post
		// To call a crossing function, we specify the `cross` keyword
		// This matches the first argument of type realm in the function itself
		err := CreatePost(cross, p.text)
		if err != nil {
			t.Fatalf("expected no error for post '%s', got %v", p.text, err)
		}
	}

	// Get the rendered page
	got := Render("")

	// Check that all posts and their authors are present in the rendered output
	for _, p := range posts {
		expectedText := p.text
		expectedAuthor := testutils.TestAddress(p.author).String() // Get the address for the author
		if !(strings.Contains(got, expectedText) && strings.Contains(got, expectedAuthor)) {
			t.Fatalf("expected render to contain text '%s' and address '%s'", expectedText, expectedAuthor)
		}
	}
}
```

Running `gno test . -v` in the `minisocial/` folder should show the tests passing:

```console
❯ gno test . -v
=== RUN   TestCreatePostSingle
--- PASS: TestCreatePostSingle (0.00s)
=== RUN   TestCreatePostMultiple
--- PASS: TestCreatePostMultiple (0.00s)
ok      .       0.87s
```

## Conclusion

Congratulations on completing your first Gno realm!
Now you're equipped with the required knowledge to venture into Gno.land.

Full code of this app can be found on the Staging network, on
[this link](https://gno.land/r/docs/minisocial).

## Bonus - resolving usernames

Let's make our MiniSocial app even better by resolving addresses to potential usernames
registered in the [Gno.land User Registry](https://gno.land/demo/users).

We can import the `gno.land/r/sys/users` realm which provides user data and use
it to try to resolve the address:

[embedmd]:# (../_assets/minisocial/types-2-bonus.gno go /\/\/ String/ $)
```go
// String stringifies a Post
func (p Post) String() string {
	out := p.text + "\n\n"

	author := p.author.String()
	// We can import and use the r/sys/users package to resolve addresses
	user, _ := users.ResolveAddress(p.author)
	if user != nil {
		// RenderLink provides a link that is clickable
		// The link goes to the user's profile page
		author = user.RenderLink()
	}

	out += ufmt.Sprintf("_by %s on %s_\n\n", author, p.createdAt.Format("02 Jan 2006, 15:04"))
	return out
}
```


