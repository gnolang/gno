# Effective Gno

This document provides advice and guidelines for writing effective Gno code.

## Formatting

### Indentation

Use 4 spaces for each level of indentation. Avoid using tabs.

```gno
func main() {
    fmt.Println("Hello, World!")
}
```

### Braces

The opening brace should be on the same line as the function declaration.

```gno
func hello() {
    // code here
}
```

## Comments

### Inline Comments

Use `//` for inline comments.

```gno
var a = 1 // This is an inline comment
```

### Block Comments

For longer descriptions, use `/* ... */`.

```gno
/*
This is a longer comment that
spans multiple lines.
*/
```

## Variables

### Declaration

Declare variables using the `var` keyword.

```gno
var x int = 10
```

You can also declare multiple variables at once:

```gno
var (
    a = 1
    b = 2
)
```

### Naming

Use camelCase for variable names.

```gno
var myVariable = "hello"
```

## Functions

### Declaration

Declare functions using the `func` keyword.

```gno
func myFunction() {
    // code here
}
```

### Parameters

Function parameters should be declared in the function signature.

```gno
func add(a int, b int) int {
    return a + b
}
```

### Return Values

Functions can return multiple values.

```gno
func divmod(a int, b int) (int, int) {
    return a / b, a % b
}
```

### Error Handling

Use errors for exception handling.

```gno
func doSomething() error {
    // if something goes wrong
    return errors.New("Something went wrong")
}
```

## Control Structures

### If Statements

If statements do not need parentheses around the condition.

```gno
if condition {
    // code here
}
```

### For Loops

Gno uses the `for` keyword for loops.

```gno
for i := 0; i < 10; i++ {
    // code here
}
```

For each loops are also supported.

```gno
for index, value := range array {
    // code here
}
```

## Structs

Structs are used to group related data together.

```gno
type Person struct {
    Name string
    Age  int
}
```

## Interfaces

Interfaces are defined using the `interface` keyword.

```gno
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

Remember, the key to effective Gno programming is practice. Write code, read code, and don't be afraid to ask for help. Happy coding!
