package errors

// Panic is a deliberate wrapper around the built-in panic function, emphasizing intentional usage.
//
// It is intended to be used in situations where there is no way to recover from an error, and where the blockchain should halt immediately.
func Panic(v any) {
	panic(v)
}
