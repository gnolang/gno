# Draft: one low-priority issue, not six

Recommendation: file this once, priority low. Do not file the six pairs separately. None of them change whether the program panics, or the index or bound in the message. They only change the sentence.

No open issue asks for these six strings to match gc. A search for `integer divide by zero`, `negative shift amount`, `nil pointer dereference`, and `slice index out of bounds` hits two closed bugs that are not this request: [gnolang/gno#1148](https://github.com/gnolang/gno/issues/1148) (division by zero could not be recovered) and [gnolang/gno#3430](https://github.com/gnolang/gno/issues/3430) (a real nil-slice panic on `len` of an array element, not a wording mismatch).

## Title

gnovm: runtime panic text does not match gc for the same events

## Summary

For several runtime panics, Gno and gc panic at the same operation and report the same numbers, but the message text differs. A differential fuzzer that compares panic strings reports these forever unless the wording is folded. Aligning the text with gc would make that class of report mean a real disagreement.

Values, control flow, and the numbers in the messages already match. This is not a semantic bug.

## Pairs

Reproduced as part of the gnodiff corpus on the same tree as upstream master `cf8ff126706cc9ba2d226bc256b3060dfa7731e2`, Go 1.25.9. Each program prints `before` and then panics. Both sides print `before`.

Division by zero. `var a, b int = 1, 0; println(a / b)`. Also every other integer width tried.

- gc: `runtime error: integer divide by zero`
- Gno: `runtime error: division by zero`

Negative shift. `var a, s int = 1, -1; println(a << s)`. The parenthesized type follows the shifted value's type (`int`, `int8`, `int16`, `int32`/`rune`, `int64`).

- gc: `runtime error: negative shift amount`
- Gno: `runtime error: negative shift amount: (-1 int)`

Index. `s := []int{10, 20}; println(s[3])`. A generated program hit the same pattern with index 5 and length 2.

- gc: `runtime error: index out of range [3] with length 2`
- Gno: `runtime error: slice index out of bounds: 3 (len=2)`

Nil pointer. `var c *cell; println(c.Twice())` with `func (c *cell) Twice() int`.

- gc: `runtime error: invalid memory address or nil pointer dereference`
- Gno: `runtime error: nil pointer dereference`

Nil slice. `var s []int; println(s[0])`.

- gc: `runtime error: index out of range [0] with length 0`
- Gno: `runtime error: nil slice index (out of bounds)`

Slice bounds with low greater than high. `s := []int{1, 2, 3}; lo, hi := 2, 1; _ = s[lo:hi]`.

- gc: `runtime error: slice bounds out of range [2:1]`
- Gno: `runtime error: invalid slice index 2 > 1`

## Expected

Use gc's sentence for each event, and keep the indexes Gno already reports. A nil-slice index should still say which index and that the length is 0, if the goal is for the strings to compare equal.

## Actual

The Gno sentences above. Behavior aside from the text matches gc on these programs.
