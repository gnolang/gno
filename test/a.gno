package main

import "std"

func Call() {
	println(std.CurrentRealm())
	println(std.PrevRealm())
}
