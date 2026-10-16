# Ternary package

Ternary operator have notoriously been absent from Go 
from its inception.

This package proposes ternary functions.

We don't advocate for their systematic use, but 
it can often prove useful when realms need to generate 
Markdown, 

## Usage
```go
import "p/grepsuzette/ternary"

func Render(path string) string {
    // display appropriate greeting
    return "# " + ternary.String(isEarly, "hi", "bye")
}
```

Another example: 

`f := ternary.Float64(useGoldenRatio, 1.618, 1.66)`

## List of functions

Most native types got a function.

Note: both branches yes/no get evaluated, contrarily to the C operator.
Please don't use this if your branches are expensive.

Functions:

* func String(cond bool, yes, no string) string
* func Int(cond bool, yes, no int) int
* func Int8(cond bool, yes, no int8) int8 
* func Int16(cond bool, yes, no int16) int16 
* func Int32(cond bool, yes, no int32) int32 
* func Int64(cond bool, yes, no int64) int64 
* func Uint(cond bool, yes, no uint) uint 
* func Uint8(cond bool, yes, no uint8) uint8 
* func Uint16(cond bool, yes, no uint16) uint16 
* func Uint32(cond bool, yes, no uint32) uint32 
* func Uint64(cond bool, yes, no uint64) uint64 
* func Float32(cond bool, yes, no float32) float32 
* func Float64(cond bool, yes, no float64) float64 
* func Rune(cond bool, yes, no rune) rune 
* func Bool(cond bool, yes, no bool) rune 
* func Address(cond bool, std.Address, std.Address) std.Address

