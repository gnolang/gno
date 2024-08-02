# Gno Doctest: Easy Code Execution and Testing

Gno Doctest is a tool that allows you to easily execute and test code blocks written in the Gno language. This tool offers a range of features, from simple code execution to complex package imports.

## Basic Usage

To use Gno Doctest, run the following command:

gno doctest -path <markdown_file_path> -run <code_block_name | "">

- `<markdown_file_path>`: Path to the markdown file containing Gno code blocks
- `<code_block_name>`: Name of the code block to run (optional)

For example, to run the code block named "print hello world" in the file "foo.md", use the following command:

gno doctest -path foo.md -run "print hello world"

## Features

### 1. Basic Code Execution

Gno Doctest can execute simple code blocks:

```go
package main

func main() {
    println("Hello, World!")
}

// Output:
// Hello, World!
```

Doctest also recognizes that a block of code is a gno. The code below outputs the same result as the example above.

```go
// @test: print hello world
package main

func main() {
    println("Hello, World!")
}

// Output:
// Hello, World!
```

Running this code will output "Hello, World!".

## 2. Using Standard Library Packages

Doctest supports automatic import and usage of standard library packages.

If run this code, doctest will automatically import the "std" package and execute the code.

```go
// @test: omit-package-declaration
func main() {
    addr := std.GetOrigCaller()
    println(addr)
}
```

The code above outputs the same result as the code below.

```go
// @test: auto-import-package
package main

import "std"

func main() {
    addr := std.GetOrigCaller()
    println(addr)
}
```

## 3. Automatic Package Import

One of the most powerful features of Gno Doctest is its ability to handle package declarations and imports automatically.

```go
func main() {
    println(math.Pi)
    println(strings.ToUpper("Hello, World"))
}
```

In this code, the math and strings packages are not explicitly imported, but Doctest automatically recognizes and imports the necessary packages.

## 4. Omitting Package Declaration

Doctest can even handle cases where the `package` declaration is omitted.

```go
// @test: omit-top-level-package-declaration
func main() {
    s := strings.ToUpper("Hello, World")
    println(s)
}

// Output:
// HELLO, WORLD
```

This code runs normally without package declaration or import statements.
Using Gno Doctest makes code execution and testing much more convenient.

You can quickly run various Gno code snippets and check the results without complex setups.

### 7. Execution Options

Doctest supports special execution options:
Ignore Option
Use the ignore tag to skip execution of a code block:

**Ignore Option**

Use the ignore tag to skip execution of a code block:

```go,ignore
// @ignore
func main() {
    println("This won't be executed")
}
```

## Conclusion

Gno Doctest simplifies the process of executing and testing Gno code snippets.
