package blog

func (b *Blog) AddModerator(addr address) error {
	return b.Authorizable.AddToAuthList(addr)
}

func (b *Blog) DeleteModerator(addr address) error {
	return b.Authorizable.DeleteFromAuthList(addr)
}

type UserResolver func(input string) (string, bool)

func CheckAddr(addr address) error {
	if !addr.IsValid() {
		return ErrInvalidCaller
	}
	return nil
}

func CheckUser(addr string, resolver UserResolver) (string, error) {
	if CheckAddr(address(addr)) != nil {
		return "", ErrInvalidCaller
	}
	if !isRegistered(addr, resolver) {
		return addr, nil
	}
	userData, _ := resolver(addr)
	return userData, nil
}

func isRegistered(addr string, resolver UserResolver) bool {
	if resolver == nil {
		return false
	}
	userData, _ := resolver(addr)
	if userData == "" {
		return false
	}
	return true
}
