package blog

import (
	"std"
)

type Permission struct { // TO DO
	Read  bool
	Write bool
}

type UserResolver func(input string) (string, bool)

func CheckAddr(addr string) error {
	if !std.Address(addr).IsValid() {
		return ErrInvalidCaller
	}
	return nil
}

func CheckUser(addr string, resolver *UserResolver) (string, error) {
	if CheckAddr(addr) != nil {
		return "", ErrInvalidCaller
	}
	if !isRegistered(addr, resolver) {
		return addr, nil
	}
	resolveUser := *resolver
	userData, _ := resolveUser(addr)
	return userData, nil
}

func isRegistered(addr string, resolver *UserResolver) bool { // TO FIX
	if resolver == nil {
		return false
	}
	resolveUser := *resolver
	userData, _ := resolveUser(addr)
	if userData == "" {
		return false
	}
	return true
}
