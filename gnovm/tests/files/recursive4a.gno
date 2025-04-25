package main

type time time.Duration

func main() {
}

// Error:
// main/files/recursive4a.gno:3:6: invalid recursive type: time -> time

// TypeCheckError:
// main/files/recursive4a.gno:3:6: invalid recursive type: time refers to itself
