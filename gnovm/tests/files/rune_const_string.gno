package main

// Converting an untyped rune constant to string gives the code point's
// UTF-8, or "�" when the constant is not a code point (Go spec,
// Conversions); an untyped rune past int32 is not one.

const c = '\U0010ffff' * '\U0010ffff'

func main() {
	println(string(c), len(string(c)), string('a'<<40), string('\U0010ffff'*1) == "\U0010ffff")
	println(string('a'), string('é'), string('a'*'b'))
}

// Output:
// � 3 � true
// a é ┢
