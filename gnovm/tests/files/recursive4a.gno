package main

type time time.Duration

func main() {
}

// Error:
// main/recursive4a.gno:3:6-24: invalid recursive type: time -> time

// TypeCheckError:
// main/recursive4a.gno:3:6: invalid recursive type: time refers to itself
