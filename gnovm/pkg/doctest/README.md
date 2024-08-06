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

### 3. Execution Options

Doctest supports special execution options:
Ignore Option
Use the ignore tag to skip execution of a code block:

**Ignore Option**

Use the ignore tag to skip execution of a code block:

```go,ignore
// @ignore
package main

func main() {
    println("This won't be executed")
}
```

## Conclusion

Gno Doctest simplifies the process of executing and testing Gno code snippets.

```go
// @test: slice
package main

type ints []int

func main() {
    a := ints{1,2,3}
    println(a)
}

// Output:
// (slice[(1 int),(2 int),(3 int)] gno.land/r/g14ch5q26mhx3jk5cxl88t278nper264ces4m8nt/run.ints)
```
