package teams

type AddressOrTeamName string

func (aotn AddressOrTeamName) IsName() bool {
	return aotn != "" && aotn[0] == '@'
}

func (aotn AddressOrTeamName) GetName() (string, bool) {
	if len(aotn) >= 2 && aotn[0] == '@' {
		return string(aotn[1:]), true
	}
	return "", false
}
