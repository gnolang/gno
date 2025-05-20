package fooo

func Foo() int64 {
    return uint64(0) // <- type mismatch
}
