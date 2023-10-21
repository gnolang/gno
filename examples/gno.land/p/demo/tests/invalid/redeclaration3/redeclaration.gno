package redeclaration

type a struct{}

func (_ a) method() int { return 1 }
