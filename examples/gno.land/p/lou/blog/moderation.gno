package blog

import (
	"std"
)

type UserResolver func(input string) (string, bool)

func CheckUser(addr string, resolver *UserResolver) (string, error) {
	if !std.Address(addr).IsValid() {
		return "", ErrInvalidCaller
	}
	if resolver == nil {
		return addr, nil
	}
	resolveUser := *resolver
	userData, _ := resolveUser(addr)
	if userData == "" {
		return addr, nil
	}
	return userData, nil
}

func IsRegistered(addr string, resolver *UserResolver) bool {
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
